/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package challenges

import (
	"bitbucket.org/kullo/server/dao"
)

type resetUsersDao interface {
	UserHasResetCode(address string) (bool, error)
	ResetCodeValid(address string, code string) (bool, error)
}

type ResetChallenge struct {
	usersDao resetUsersDao
}

func NewResetChallenge() ResetChallenge {
	return ResetChallenge{usersDao: &dao.Users{}}
}

func (self *ResetChallenge) Name() string {
	return "reset"
}

func (self *ResetChallenge) ChallengeNecessary(address string) (bool, error) {
	// reset codes cannot be used on an address that doesn't exist
	return false, nil
}

func (self *ResetChallenge) FillChallenge(challenge *Challenge, address string,
	userExists bool, isLocalAddress bool) error {

	if !userExists {
		return ErrNotApplicable
	}
	hasResetCodeResult, err := self.usersDao.UserHasResetCode(address)
	if err != nil {
		return err
	}
	if !hasResetCodeResult {
		return ErrNotApplicable
	}

	challenge.Text = "Please enter the reset code for address '" + address + "'."
	return nil
}

func (self *ResetChallenge) CheckChallenge(clientAnswer *ChallengeClientAnswer,
	userExists bool, isLocalAddress bool) (bool, error) {

	if !userExists {
		return false, nil
	}

	return self.usersDao.ResetCodeValid(clientAnswer.Address, clientAnswer.ChallengeAnswer)
}
