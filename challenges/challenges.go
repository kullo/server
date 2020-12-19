/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package challenges

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"time"
)

var ErrUserExists = errors.New("User already exists")

type Challenge struct {
	Type      string `json:"type"`
	User      string `json:"user"`
	Timestamp uint64 `json:"timestamp"`
	Text      string `json:"text"`
}

type ChallengeClientAnswer struct {
	Address         string    `json:"address"`
	Challenge       Challenge `json:"challenge"`
	ChallengeAuth   string    `json:"challengeAuth"`
	ChallengeAnswer string    `json:"challengeAnswer"`
}

type ChallengeReply struct {
	Challenge     Challenge `json:"challenge"`
	ChallengeAuth string    `json:"challengeAuth"`
}

var ErrNotApplicable = errors.New("Challenge is not applicable to given user")
var ErrNoApplicableChallengeTypeFound = errors.New("No applicable challenge type has been found")

type ChallengeType interface {
	// Returns the challenge name as used in Challenge.Type
	Name() string

	// Is this challenge required under the assumption that the address doesn't
	// exist and challenges are optional?
	ChallengeNecessary(address string) (bool, error)

	// Fill the fields of the challenge being passed in
	FillChallenge(challenge *Challenge, address string, userExists bool, isLocalAddress bool) error

	// Check the validity of the ChallengeAnswer field
	CheckChallenge(clientAnswer *ChallengeClientAnswer, userExists bool, isLocalAddress bool) (bool, error)
}

var resetChallenge = NewResetChallenge()
var reservationChallenge = NewReservationChallenge()
var codeChallenge = NewCodeChallenge()
var blockedChallenge = NewBlockedChallenge()

// The first applicable challenge is used, so order matters! Rules:
// * Reservation must precede code, so that reserved addresses cannot be taken.
// * There must be at least one catch-all challenge type for userExists == false if
//   challengesAreOptional == false.
// * Blocked must be last, because it blocks all attempts at registration...
var challengeTypes = []ChallengeType{
	&resetChallenge,
	&reservationChallenge,
	&codeChallenge,
	&blockedChallenge,
}

const challengesAreOptional bool = true

func challengeNecessary(address string, userExists bool, isLocalAddress bool) (bool, error) {
	if !challengesAreOptional || userExists || !isLocalAddress {
		return true, nil
	}

	for _, challengeType := range challengeTypes {
		necessary, err := challengeType.ChallengeNecessary(address)
		if err != nil {
			return true, err
		}
		if necessary {
			return true, nil
		}
	}

	// only return false if
	// * challenges are optional,
	// * the user doesn't exist, and
	// * no challenge declares itself necessary
	return false, nil
}

var challengeKey = []byte("UPcYZJbOTcwziuPxLzKMwECqyDoFsJYhJDJilPgQRnUOvucItYKaoIhNIULuYUDr")

func serializeChallenge(challenge *Challenge) string {
	return challenge.Type +
		"|||" + challenge.User +
		"|||" + strconv.FormatUint(challenge.Timestamp, 10) +
		"|||" + challenge.Text
}

func createChallengeAuth(challenge *Challenge) string {
	message := []byte(serializeChallenge(challenge))
	mac := hmac.New(sha256.New, challengeKey)
	mac.Write(message)
	outBytes := mac.Sum(nil)
	return hex.EncodeToString(outBytes)
}

func CreateChallenge(address string, userExists bool, isLocalAddress bool) (*ChallengeReply, error) {
	// return nil if no challenge is necessary
	necessary, err := challengeNecessary(address, userExists, isLocalAddress)
	if err != nil {
		return nil, err
	}
	if !necessary {
		return nil, nil
	}

	challenge := Challenge{}
	challenge.User = address
	challenge.Timestamp = uint64(time.Now().Unix())

	// fill challenge with first matching challenge type
	isChallengeFilled := false
	for _, challengeType := range challengeTypes {
		err := challengeType.FillChallenge(&challenge, address, userExists, isLocalAddress)
		switch err {
		case nil:
			challenge.Type = challengeType.Name()
			isChallengeFilled = true
		case ErrNotApplicable:
			// go to next challenge type
		default:
			return nil, err
		}

		if isChallengeFilled {
			break
		}
	}
	if !isChallengeFilled {
		if userExists {
			return nil, ErrUserExists
		} else {
			return nil, ErrNoApplicableChallengeTypeFound
		}
	}

	challengeReply := &ChallengeReply{challenge, createChallengeAuth(&challenge)}
	return challengeReply, nil
}

func CheckChallenge(clientAnswer *ChallengeClientAnswer, userExists bool, isLocalAddress bool) (bool, error) {
	// check whether challenge is necessary at all
	necessary, err := challengeNecessary(clientAnswer.Address, userExists, isLocalAddress)
	if err != nil {
		return false, err
	}
	if !necessary {
		// no challenge necessary -> immediate success
		return true, nil
	}

	// challenge has been generated for the user that tries to register
	if clientAnswer.Challenge.User != clientAnswer.Address {
		return false, nil
	}

	// challenge not older than 15mins
	if clientAnswer.Challenge.Timestamp < (uint64(time.Now().Unix()) - 15*60) {
		return false, nil
	}

	// auth okay
	authMacHexExpected := []byte(createChallengeAuth(&clientAnswer.Challenge))
	authMacHexActual := []byte(clientAnswer.ChallengeAuth)

	// constant time compare of raw binary strings.
	//
	// hmac.Equal() expects two raw MACs but gets byte strings of hashes of
	// MACs. Fortunately it's source code is simple enough to compare
	// arbitrary byte strings: http://golang.org/src/pkg/crypto/hmac/hmac.go
	if !hmac.Equal(authMacHexExpected, authMacHexActual) {
		return false, nil
	}

	// check challenge with first matching challenge type
	for _, challengeType := range challengeTypes {
		if challengeType.Name() == clientAnswer.Challenge.Type {
			return challengeType.CheckChallenge(clientAnswer, userExists, isLocalAddress)
		}
	}

	// no matching ChallengeType has been found
	if userExists {
		return false, ErrUserExists
	} else {
		return false, nil
	}
}
