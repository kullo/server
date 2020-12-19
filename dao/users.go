/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package dao

import (
	"database/sql"
	"errors"
	"strings"

	"bitbucket.org/kullo/server/dbconn"
)

type UsersEntry struct {
	ID               uint32 `json:"id"`
	AcceptedTerms    string `json:"acceptedTerms"`
	PlanName         string `json:"planName"`
	StorageQuota     uint64 `json:"storageQuota"`
	ResetCode        string `json:"-"`
	WebloginUsername string `json:"-"`
	WebloginSecret   string `json:"-"`
	Language         string `json:"-"`
}

type AddressesEntry struct {
	Address          string `json:"address"`
	RegistrationCode string `json:"-"`
}

type Users struct {
}

var ErrAddressAlreadyExists = errors.New("dao: address already exists.")

func (dao *Users) UserExists(address string) (bool, error) {
	var exists bool
	err := dbconn.GetConn().
		QueryRow("SELECT count(id) > 0 "+
			"FROM addresses "+
			"WHERE address=$1 ",
			address).
		Scan(&exists)
	return exists, err
}

func (dao *Users) UserExistsAndActive(address string) (bool, error) {
	var exists bool
	err := dbconn.GetConn().
		QueryRow("SELECT count(u.id) > 0 "+
			"FROM users u JOIN addresses a ON u.id = a.user_id "+
			"WHERE u.disabled=FALSE AND a.address=$1 ",
			address).
		Scan(&exists)
	return exists, err
}

func (dao *Users) RegistrationCodeUsed(code string) (bool, error) {
	var exists bool
	err := dbconn.GetConn().
		QueryRow("SELECT count(id) > 0 "+
			"FROM addresses "+
			"WHERE registration_code=$1 ",
			code).
		Scan(&exists)
	return exists, err
}

func (dao *Users) UserHasResetCode(address string) (bool, error) {
	var hasCode bool
	err := dbconn.GetConn().
		QueryRow("SELECT length(u.reset_code) > 0 "+
			"FROM users u JOIN addresses a ON u.id = a.user_id "+
			"WHERE a.address=$1 ",
			address).
		Scan(&hasCode)
	return hasCode, err
}

func (dao *Users) ResetCodeValid(address string, code string) (bool, error) {
	var valid bool
	err := dbconn.GetConn().
		QueryRow("SELECT count(u.id) > 0 "+
			"FROM users u JOIN addresses a ON u.id = a.user_id "+
			"WHERE a.address=$1 AND u.reset_code=$2 ",
			address, code).
		Scan(&valid)
	return valid, err
}

func (dao *Users) getPlanId(transaction *sql.Tx, planName string) (uint32, error) {
	var id uint32
	err := transaction.QueryRow(
		"SELECT id FROM plans WHERE name = $1 ",
		planName).
		Scan(&id)
	return id, err
}

func (dao *Users) getDefaultPlanName(address string) string {
	if strings.HasSuffix(address, "#kullo.net") || strings.HasSuffix(address, "#kullo.test") {
		return "Free"
	} else {
		return "Professional"
	}
}

func (dao *Users) InsertEntry(entry *AddressesEntry, acceptedTerms string) error {
	transaction, err := dbconn.GetConn().Begin()
	if err != nil {
		return err
	}

	var exists bool
	err = transaction.QueryRow(
		"SELECT count(id) > 0 FROM addresses WHERE address=$1",
		entry.Address).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		err = transaction.Rollback()
		if err != nil {
			return err
		}
		return ErrAddressAlreadyExists
	}

	planId, err := dao.getPlanId(transaction, dao.getDefaultPlanName(entry.Address))
	if err != nil {
		return err
	}

	var userId uint32
	err = transaction.QueryRow(
		"INSERT INTO users (accepted_terms, plan_id) VALUES ($1, $2) RETURNING id",
		acceptedTerms, planId).Scan(&userId)
	if err != nil {
		return err
	}
	_, err = transaction.Exec(
		"INSERT INTO addresses (user_id, address, registration_code) "+
			"VALUES ($1, $2, $3)",
		userId, entry.Address, entry.RegistrationCode)
	if err != nil {
		return err
	}

	return transaction.Commit()
}

func (dao *Users) GetEntry(address string) (*UsersEntry, error) {
	entry := &UsersEntry{}
	err := dbconn.GetConn().
		QueryRow(
			"SELECT u.id, u.reset_code, u.accepted_terms, "+
				"p.name, p.storage_quota, "+
				"u.weblogin_username, u.weblogin_secret, "+
				"u.language "+
				"FROM users u, plans p, addresses a "+
				"WHERE u.plan_id=p.id AND u.id=a.user_id AND a.address=$1", address).
		Scan(&entry.ID, &entry.ResetCode, &entry.AcceptedTerms,
			&entry.PlanName, &entry.StorageQuota,
			&entry.WebloginUsername, &entry.WebloginSecret,
			&entry.Language)
	return entry, err
}

func (dao *Users) Reset(address string) error {
	transaction, err := dbconn.GetConn().Begin()
	if err != nil {
		return err
	}

	var userId uint64
	err = transaction.QueryRow(
		"SELECT user_id FROM addresses WHERE address=$1",
		address).Scan(&userId)
	if err != nil {
		return err
	}

	// delete reset code
	_, err = transaction.Exec(
		"UPDATE users SET reset_code='' WHERE id=$1",
		userId)
	if err != nil {
		return err
	}

	// delete all messages
	_, err = transaction.Exec(
		"DELETE FROM messages WHERE user_id=$1",
		userId)
	if err != nil {
		return err
	}

	// delete all profile information
	_, err = transaction.Exec(
		"DELETE FROM profile WHERE user_id=$1",
		userId)
	if err != nil {
		return err
	}

	// delete all encryption keys
	_, err = transaction.Exec(
		"DELETE FROM keys_asymm "+
			"WHERE user_id=$1 AND key_type='enc'", //FIXME magic constant
		userId)
	if err != nil {
		return err
	}

	// delete private signature verification keys
	_, err = transaction.Exec(
		"UPDATE keys_asymm SET privkey='' "+
			"WHERE user_id=$1 AND key_type='sig'", //FIXME magic constant
		userId)
	if err != nil {
		return err
	}

	return transaction.Commit()
}

func (dao *Users) UpdateLatestLoginTimestamp(address string) error {
	_, err := dbconn.GetConn().Exec(
		"UPDATE users u "+
			"SET last_login = now() "+
			"FROM addresses a "+
			"WHERE u.id = a.user_id AND a.address = $1 ",
		address)
	return err
}

func (dao *Users) UpdateLatestLoginTimestampAndLanguage(address string, language string) error {
	_, err := dbconn.GetConn().Exec(
		"UPDATE users u "+
			"SET last_login = now(), language = $1 "+
			"FROM addresses a "+
			"WHERE u.id = a.user_id AND a.address = $2 ",
		language, address)
	return err
}
