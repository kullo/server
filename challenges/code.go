/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package challenges

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"

	"bitbucket.org/kullo/server/dao"
	"golang.org/x/crypto/hkdf"
)

var ErrWrongOkmLength = errors.New("challenges: wrong OKM length")

type regCodeUsersDao interface {
	RegistrationCodeUsed(code string) (bool, error)
}

type CodeChallenge struct {
	usersDao regCodeUsersDao
}

func NewCodeChallenge() CodeChallenge {
	return CodeChallenge{usersDao: &dao.Users{}}
}

func (self *CodeChallenge) Name() string {
	return "code"
}

func (self *CodeChallenge) ChallengeNecessary(address string) (bool, error) {
	// this challenge is not necessary if challenges are optional
	return false, nil
}

func (self *CodeChallenge) FillChallenge(challenge *Challenge, address string,
	userExists bool, isLocalAddress bool) error {

	// the code challenge only applies to local addresses
	if userExists || !isLocalAddress {
		return ErrNotApplicable
	}

	challenge.Text = "Please enter an invite code."
	return nil
}

func (self *CodeChallenge) CheckChallenge(clientAnswer *ChallengeClientAnswer,
	userExists bool, isLocalAddress bool) (bool, error) {

	// the code challenge only applies to local addresses
	if userExists || !isLocalAddress {
		return false, nil
	}

	return self.validateCode(clientAnswer.ChallengeAnswer)
}

// ### begin private stuff ###

const maxCodeId = 4095

var codeRegex = regexp.MustCompile("[0-9a-f]{16}([0-9a-f]{3,4})")
var codeMasterSecret = []byte("vQTRxqOiUfBGDjbYXcIEoVwRedEUQyRXSOfJQVqohQYpLdDStVYViibXplMCKOhJ")

func (self *CodeChallenge) hkdf512(inputKeyingMaterial []byte, length int, info []byte) ([]byte, error) {
	hkdf := hkdf.New(sha512.New, inputKeyingMaterial, nil, info)

	outputKeyingMaterial := make([]byte, length)
	n, err := io.ReadFull(hkdf, outputKeyingMaterial)
	if err != nil {
		return nil, err
	}
	if n != len(outputKeyingMaterial) {
		return nil, ErrWrongOkmLength
	}
	return outputKeyingMaterial, nil
}

func (self *CodeChallenge) generateCode(id int64) (string, error) {
	idStr := fmt.Sprintf("%03x", id)

	authBlob, err := self.hkdf512(codeMasterSecret, 8, []byte(idStr))
	if err != nil {
		return "", err
	}
	code := hex.EncodeToString(authBlob) + idStr
	return code, nil
}

func (self *CodeChallenge) validateCode(code string) (bool, error) {
	matches := codeRegex.FindStringSubmatch(code)
	if len(matches) != 2 {
		return false, nil
	}
	idStr := matches[1]

	id, err := strconv.ParseInt(idStr, 16, 64)
	if err != nil {
		return false, nil // don't return err, this is a client-side error
	}
	if id < 0 || id > maxCodeId {
		return false, nil
	}

	expectedCode, err := self.generateCode(id)
	if err != nil {
		return false, err
	}
	if code != expectedCode {
		return false, nil
	}

	used, err := self.usersDao.RegistrationCodeUsed(code)
	if err != nil {
		return false, err
	}
	if used {
		return false, nil
	}

	return true, nil
}
