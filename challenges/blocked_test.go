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

func makeBlockedUut() *BlockedChallenge {
	uut := NewBlockedChallenge()
	return &uut
}

func TestBlockedName(t *testing.T) {
	uut := makeBlockedUut()
	name := uut.Name()
	if name != "blocked" {
		t.Error("Name is", name)
	}
}

func TestBlockedChallengeNecessary(t *testing.T) {
	uut := makeBlockedUut()

	// necessary case (challenge reserved address)
	necessary, err := uut.ChallengeNecessary("any.address#some.domain")
	if err != nil {
		t.Fatal("ChallengeNecessary failed:", err)
	}
	if necessary {
		t.Error("this should not be necessary")
	}

	// unnecessary case (challenge non-reserved address)
	necessary, err = uut.ChallengeNecessary("someone#kullo.net")
	if err != nil {
		t.Fatal("ChallengeNecessary failed:", err)
	}
	if necessary {
		t.Error("this should not be necessary")
	}
}

func TestBlockedFillChallenge(t *testing.T) {
	uut := makeBlockedUut()
	unmodifiedChallenge := Challenge{}

	// for blocked, it is of no importance whether address is local
	isLocalAddress := true

	// user doesn't exist -> success
	userExists := false
	challenge := Challenge{}
	err := uut.FillChallenge(&challenge, addressWithReservation, userExists, isLocalAddress)
	if err != nil {
		t.Fatal("FillChallenge failed:", err)
	}
	if challenge.Text == "" {
		t.Error("Text is empty")
	}

	// user does exist -> fail
	userExists = true
	challenge = Challenge{}
	err = uut.FillChallenge(&challenge, "nobody#kullo.net", userExists, isLocalAddress)
	if err != ErrNotApplicable {
		t.Fatal("FillChallenge didn't fail with ErrNotApplicable:", err)
	}
	if challenge != unmodifiedChallenge {
		t.Error("Challenge has been modified")
	}
}

func TestBlockedCheckChallenge(t *testing.T) {
	uut := makeBlockedUut()

	// for blocked, these are of no importance
	userExists := false
	isLocalAddress := true

	// always fail
	clientAnswer := ChallengeClientAnswer{}
	ok, err := uut.CheckChallenge(&clientAnswer, userExists, isLocalAddress)
	if ok != false || err != nil {
		t.Error("should fail")
	}
}
