/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package challenges

type BlockedChallenge struct {
}

func NewBlockedChallenge() BlockedChallenge {
	return BlockedChallenge{}
}

func (self *BlockedChallenge) Name() string {
	return "blocked"
}

func (self *BlockedChallenge) ChallengeNecessary(address string) (bool, error) {
	// this challenge is not necessary if challenges are optional
	return false, nil
}

func (self *BlockedChallenge) FillChallenge(challenge *Challenge, address string,
	userExists bool, isLocalAddress bool) error {

	if userExists {
		return ErrNotApplicable
	}

	challenge.Text = "The address '" + address + "' cannot be registered."
	return nil
}

func (self *BlockedChallenge) CheckChallenge(clientAnswer *ChallengeClientAnswer,
	userExists bool, isLocalAddress bool) (bool, error) {

	// this challenge cannot be successfully answered because it indicates a blocked address
	return false, nil
}
