/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package challenges

import (
	"errors"
	"testing"
)

const (
	regCode0      string = "537cc9ed8af590ae000"
	regCode1      string = "2f3e4110df958da7001"
	regCode2      string = "bd48ea92eec92872002"
	regCode0xfff  string = "d94ef0ad74b0d6b3fff"
	regCode0x1000 string = "8428db8d3c43f76e1000"
)

type codeUsersDaoStub struct{}

func (self *codeUsersDaoStub) RegistrationCodeUsed(code string) (bool, error) {
	switch code {
	case regCode1:
		return false, nil
	case regCode2:
		return true, nil
	default:
		return false, errors.New("Some error")
	}
}

func makeCodeUut() *CodeChallenge {
	uut := NewCodeChallenge()
	uut.usersDao = &codeUsersDaoStub{}
	return &uut
}

func TestCodeName(t *testing.T) {
	uut := makeCodeUut()
	name := uut.Name()
	if name != "code" {
		t.Error("Name is", name)
	}
}

func TestCodeChallengeNecessary(t *testing.T) {
	uut := makeCodeUut()
	necessary, err := uut.ChallengeNecessary("someone#kullo.net")
	if err != nil {
		t.Fatal("ChallengeNecessary failed:", err)
	}
	if necessary {
		t.Error("this should not be necessary")
	}
}

func TestCodeFillChallenge(t *testing.T) {
	uut := makeCodeUut()

	// user doesn't exist -> success
	userExists := false
	isLocalAddress := true

	challenge := Challenge{}
	err := uut.FillChallenge(&challenge, "someone#kullo.net", userExists, isLocalAddress)
	if err != nil {
		t.Fatal("FillChallenge failed:", err)
	}
	if challenge.Text == "" {
		t.Error("Text is empty")
	}

	// user does exist -> fail
	userExists = true
	isLocalAddress = true

	challenge = Challenge{}
	err = uut.FillChallenge(&challenge, "someone#kullo.net", userExists, isLocalAddress)
	if err != ErrNotApplicable {
		t.Fatal("FillChallenge didn't fail with ErrNotApplicable:", err)
	}
	unmodifiedChallenge := Challenge{}
	if challenge != unmodifiedChallenge {
		t.Error("Challenge has been modified")
	}

	// address is non-local -> fail
	userExists = false
	isLocalAddress = false

	challenge = Challenge{}
	err = uut.FillChallenge(&challenge, "someone#kullo.net", userExists, isLocalAddress)
	if err != ErrNotApplicable {
		t.Fatal("FillChallenge didn't fail with ErrNotApplicable:", err)
	}
	unmodifiedChallenge = Challenge{}
	if challenge != unmodifiedChallenge {
		t.Error("Challenge has been modified")
	}
}

func TestCodeCheckChallenge(t *testing.T) {
	uut := makeCodeUut()

	// user exists -> fail
	userExists := true
	isLocalAddress := true

	clientAnswer := ChallengeClientAnswer{}
	ok, err := uut.CheckChallenge(&clientAnswer, userExists, isLocalAddress)
	if ok != false || err != nil {
		t.Error("should fail for existing user")
	}

	// non-local user -> fail
	userExists = false
	isLocalAddress = false

	clientAnswer = ChallengeClientAnswer{}
	ok, err = uut.CheckChallenge(&clientAnswer, userExists, isLocalAddress)
	if ok != false || err != nil {
		t.Error("should fail for non-local user")
	}

	// local user doesn't exist, bad code -> fail
	userExists = false
	isLocalAddress = true

	clientAnswer = ChallengeClientAnswer{ChallengeAnswer: "invalid"}
	ok, err = uut.CheckChallenge(&clientAnswer, userExists, isLocalAddress)
	if ok != false || err != nil {
		t.Error("should fail for invalid code")
	}

	// local user doesn't exist, good code -> success
	userExists = false
	isLocalAddress = true

	clientAnswer = ChallengeClientAnswer{ChallengeAnswer: regCode1}
	ok, err = uut.CheckChallenge(&clientAnswer, userExists, isLocalAddress)
	if ok != true || err != nil {
		t.Error("should succeed for valid code")
	}
}

func expectCode(t *testing.T, id int64, expected string) {
	uut := makeCodeUut()
	code, err := uut.generateCode(id)
	if err != nil {
		t.Fatal("generateCode failed")
	}
	if code != expected {
		t.Error("Code", id, "is", code)
	}
}

func TestCodeGenerateCode(t *testing.T) {
	expectCode(t, 0, regCode0)
	expectCode(t, 1, regCode1)
	expectCode(t, 2, regCode2)

	expectCode(t, 0xfff, regCode0xfff)
	expectCode(t, 0x1000, regCode0x1000)
}

func expectCodeValidateResult(t *testing.T, code string, expected bool, errExpected bool, errmsg string) {
	uut := makeCodeUut()
	ok, err := uut.validateCode(code)
	if !errExpected && (err != nil) {
		t.Fatal("error in validateCode")
	}
	if errExpected && (err == nil) {
		t.Fatal("error in validateCode expected")
	}
	if ok != expected {
		t.Error(errmsg)
	}
}

func TestCodeValidateCode(t *testing.T) {
	// valid
	expectCodeValidateResult(t, regCode1, true, false, "code 1 should be valid")

	// non-hex char in auth part
	wrongChar := "x" + regCode1[1:]
	expectCodeValidateResult(t, wrongChar, false, false, "wrongChar should be invalid")

	// non-hex char in id part
	badId := regCode1[:len(regCode1)-1] + "x"
	expectCodeValidateResult(t, badId, false, false, "badId should be invalid")

	// invalid auth
	invalidAuth := "0123456789abcdef000"
	expectCodeValidateResult(t, invalidAuth, false, false, "invalidAuth should be invalid")

	// code already used
	expectCodeValidateResult(t, regCode2, false, false, "code 2 should be invalid")

	// error on usersDao.RegistrationCodeUsed
	expectCodeValidateResult(t, regCode0xfff, false, true, "code 0xfff should be invalid")

	// id > maxCodeId
	expectCodeValidateResult(t, regCode0x1000, false, false, "code 0x1000 should be invalid")
}
