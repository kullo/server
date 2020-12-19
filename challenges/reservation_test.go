/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package challenges

import (
	"bytes"
	"io"
	"testing"
)

const (
	addressWithReservation string = "has.reservation#kullo.net"
	validReservationCode   string = "reservation_code"
)

// Fills a bytes.Buffer with test data and extends it to satisfy io.ReadCloser
type preregFileStub struct {
	*bytes.Buffer
}

func newPreregFileStub() (io.ReadCloser, error) {
	stub := &preregFileStub{
		bytes.NewBufferString(
			addressWithReservation + ",,,," + validReservationCode + ",",
		),
	}
	return stub, nil
}

func (self *preregFileStub) Close() error {
	return nil
}

func makeReservationUut() *ReservationChallenge {
	uut := NewReservationChallenge()
	uut.openPreregFile = newPreregFileStub
	return &uut
}

func TestReservationName(t *testing.T) {
	uut := makeReservationUut()
	name := uut.Name()
	if name != "reservation" {
		t.Error("Name is", name)
	}
}

func TestReservationChallengeNecessary(t *testing.T) {
	uut := makeReservationUut()

	// necessary case (challenge 	reserved address)
	necessary, err := uut.ChallengeNecessary(addressWithReservation)
	if err != nil {
		t.Fatal("ChallengeNecessary failed:", err)
	}
	if !necessary {
		t.Error("this should be necessary")
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

func TestReservationFillChallenge(t *testing.T) {
	uut := makeReservationUut()
	unmodifiedChallenge := Challenge{}

	// for reservation, it is of no importance whether address is local
	isLocalAddress := true

	// user doesn't exist, but has a reservation code -> success
	userExists := false
	challenge := Challenge{}
	err := uut.FillChallenge(&challenge, addressWithReservation, userExists, isLocalAddress)
	if err != nil {
		t.Fatal("FillChallenge failed:", err)
	}
	if challenge.Text == "" {
		t.Error("Text is empty")
	}

	// user doesn't exist and doesn't have a reservation code -> fail
	userExists = false
	challenge = Challenge{}
	err = uut.FillChallenge(&challenge, "someone#kullo.net", userExists, isLocalAddress)
	if err != ErrNotApplicable {
		t.Fatal("FillChallenge didn't fail with ErrNotApplicable:", err)
	}
	if challenge != unmodifiedChallenge {
		t.Error("Challenge has been modified")
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

func TestReservationCheckChallenge(t *testing.T) {
	uut := makeReservationUut()

	// for reservation, it is of no importance whether address is local
	isLocalAddress := true

	// user exists -> fail
	clientAnswer := ChallengeClientAnswer{}
	userExists := true
	ok, err := uut.CheckChallenge(&clientAnswer, userExists, isLocalAddress)
	if ok != false || err != nil {
		t.Error("should fail for existing user")
	}

	// user doesn't exist, bad code -> fail
	clientAnswer = ChallengeClientAnswer{
		Address:         addressWithReservation,
		ChallengeAnswer: "invalid",
	}
	userExists = false
	ok, err = uut.CheckChallenge(&clientAnswer, userExists, isLocalAddress)
	if ok != false || err != nil {
		t.Error("should fail for invalid code")
	}

	// user doesn't exist, good code -> success
	clientAnswer = ChallengeClientAnswer{
		Address:         addressWithReservation,
		ChallengeAnswer: validReservationCode,
	}
	userExists = false
	ok, err = uut.CheckChallenge(&clientAnswer, userExists, isLocalAddress)
	if ok != true || err != nil {
		t.Error("should succeed for valid code")
	}
}
