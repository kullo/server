/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package challenges

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
)

var ErrReservationListIntegrity = errors.New("Data integrity error in CSV: code must not be empty.")

type openFunc func() (io.ReadCloser, error)

type ReservationChallenge struct {
	openPreregFile openFunc
}

func NewReservationChallenge() ReservationChallenge {
	return ReservationChallenge{openPreregFile: openPreregFile}
}

func (self *ReservationChallenge) Name() string {
	return "reservation"
}

func (self *ReservationChallenge) ChallengeNecessary(address string) (bool, error) {
	// this challenge is necessary iff the address has been reserved
	return self.isReserved(address)
}

func (self *ReservationChallenge) FillChallenge(challenge *Challenge, address string,
	userExists bool, isLocalAddress bool) error {

	if userExists {
		return ErrNotApplicable
	}
	isResv, err := self.isReserved(address)
	if err != nil {
		return err
	}
	if !isResv {
		return ErrNotApplicable
	}

	challenge.Text = "Please enter a reservation code for address '" + address + "'."
	return nil
}

func (self *ReservationChallenge) CheckChallenge(clientAnswer *ChallengeClientAnswer,
	userExists bool, isLocalAddress bool) (bool, error) {

	if userExists {
		return false, nil
	}

	codeExpected, err := self.getReservationCode(clientAnswer.Address)
	if err != nil {
		return false, err
	}
	return codeExpected == clientAnswer.ChallengeAnswer, nil
}

// ### begin private stuff ###

func openPreregFile() (io.ReadCloser, error) {
	return os.Open("config/preregistrations.csv")
}

func (self *ReservationChallenge) getReservationCode(address string) (string, error) {
	file, err := self.openPreregFile()
	if err != nil {
		return "", err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}

		rowAddress := row[0]
		rowCode := row[4]

		if rowCode == "" {
			return "", ErrReservationListIntegrity
		}

		if rowAddress == address { // address matches
			return rowCode, nil
		}
	}

	return "", nil // not found
}

func (self *ReservationChallenge) isReserved(address string) (bool, error) {
	code, err := self.getReservationCode(address)
	return (code != ""), err
}
