/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package challenges

import (
	"testing"
)

const (
	userWithResetCode string = "has.reset.code#kullo.net"
	validResetCode    string = "valid_reset_code"
)

type resetUsersDaoStub struct{}

func (self *resetUsersDaoStub) UserHasResetCode(address string) (bool, error) {
	return address == userWithResetCode, nil
}

func (self *resetUsersDaoStub) ResetCodeValid(address string, code string) (bool, error) {
	return (address == userWithResetCode) && (code == validResetCode), nil
}

func makeResetUut() *ResetChallenge {
	uut := NewResetChallenge()
	uut.usersDao = &resetUsersDaoStub{}
	return &uut
}

func TestResetName(t *testing.T) {
	uut := makeResetUut()
	name := uut.Name()
	if name != "reset" {
		t.Error("Name is", name)
	}
}

func TestResetChallengeNecessary(t *testing.T) {
	uut := makeResetUut()
	necessary, err := uut.ChallengeNecessary("someone#kullo.net")
	if err != nil {
		t.Fatal("ChallengeNecessary failed:", err)
	}
	if necessary {
		t.Error("this should not be necessary")
	}
}

func TestResetFillChallenge(t *testing.T) {
	uut := makeResetUut()
	unmodifiedChallenge := Challenge{}

	// for reset, it is of no importance whether address is local
	isLocalAddress := true

	// user exists -> success
	userExists := true

	challenge := Challenge{}
	err := uut.FillChallenge(&challenge, userWithResetCode, userExists, isLocalAddress)
	if err != nil {
		t.Fatal("FillChallenge failed:", err)
	}
	if challenge.Text == "" {
		t.Error("Text is empty")
	}

	// user doesn't have a reset code -> fail
	userExists = true

	challenge = Challenge{}
	err = uut.FillChallenge(&challenge, "someone#kullo.net", userExists, isLocalAddress)
	if err != ErrNotApplicable {
		t.Fatal("FillChallenge didn't fail with ErrNotApplicable:", err)
	}
	if challenge != unmodifiedChallenge {
		t.Error("Challenge has been modified")
	}

	// user doesn't exist -> fail
	userExists = false

	challenge = Challenge{}
	err = uut.FillChallenge(&challenge, "someone#kullo.net", userExists, isLocalAddress)
	if err != ErrNotApplicable {
		t.Fatal("FillChallenge didn't fail with ErrNotApplicable:", err)
	}
	if challenge != unmodifiedChallenge {
		t.Error("Challenge has been modified")
	}
}

func TestResetCheckChallenge(t *testing.T) {
	uut := makeResetUut()

	// for reset, it is of no importance whether address is local
	isLocalAddress := true

	// user doesn't exist -> fail
	clientAnswer := ChallengeClientAnswer{}
	userExists := false
	ok, err := uut.CheckChallenge(&clientAnswer, userExists, isLocalAddress)
	if ok != false || err != nil {
		t.Error("should fail for non-existing user")
	}

	// user doesn't have a reset code -> fail
	clientAnswer = ChallengeClientAnswer{
		Address:         "user.without.reset.code#kullo.net",
		ChallengeAnswer: validResetCode,
	}
	userExists = true
	ok, err = uut.CheckChallenge(&clientAnswer, userExists, isLocalAddress)
	if ok != false || err != nil {
		t.Error("should fail for user without reset code")
	}

	// reset code doesn't match -> fail
	clientAnswer = ChallengeClientAnswer{
		Address:         userWithResetCode,
		ChallengeAnswer: "invalid",
	}
	userExists = true
	ok, err = uut.CheckChallenge(&clientAnswer, userExists, isLocalAddress)
	if ok != false || err != nil {
		t.Error("should fail for invalid code")
	}

	// reset code matches -> success
	clientAnswer = ChallengeClientAnswer{
		Address:         userWithResetCode,
		ChallengeAnswer: validResetCode,
	}
	userExists = true
	ok, err = uut.CheckChallenge(&clientAnswer, userExists, isLocalAddress)
	if ok != true || err != nil {
		t.Error("should succeed for valid code")
	}
}
