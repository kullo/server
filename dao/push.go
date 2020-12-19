/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package dao

import (
	"database/sql"

	"bitbucket.org/kullo/server/dbconn"
)

type NotificationsGcmEntry struct {
	RegistrationToken string `json:"registrationToken"`
	Environment       string `json:"environment"`
}

type NotificationsGcm struct {
}

func (dao *NotificationsGcm) GetTokens(address string) ([]NotificationsGcmEntry, error) {
	rows, err := dbconn.GetConn().
		Query("SELECT ng.registration_token, ng.environment "+
			"FROM notifications_gcm ng JOIN addresses a ON ng.user_id=a.user_id "+
			"WHERE a.address=$1", address)
	if err != nil {
		return nil, err
	}

	entries := []NotificationsGcmEntry{}
	entry := NotificationsGcmEntry{}
	for rows.Next() {
		err = rows.Scan(&entry.RegistrationToken, &entry.Environment)
		if err != nil {
			rows.Close()
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (dao *NotificationsGcm) InsertEntry(address string, entry *NotificationsGcmEntry) error {
	tx, err := dbconn.GetConn().Begin()
	if err != nil {
		return err
	}

	// delete other tokens with the same instance ID (e.g. old GCM tokens when we use FCM)
	query := "DELETE FROM notifications_gcm " +
		"WHERE left(registration_token, 12) = $1 || ':' AND registration_token != $2 "
	instanceId := entry.RegistrationToken[:11]
	_, err = tx.Exec(query, instanceId, entry.RegistrationToken)
	if err != nil {
		tx.Rollback()
		return err
	}

	// check whether this token already exists
	query = "SELECT a.address, ng.environment FROM notifications_gcm ng, addresses a " +
		"WHERE ng.user_id=a.user_id AND ng.registration_token=$1 " +
		"LIMIT 1"
	var existingAddress string
	var existingEnv string
	err = tx.QueryRow(query, entry.RegistrationToken).Scan(&existingAddress, &existingEnv)
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return err
	}

	if err == sql.ErrNoRows {
		// insert the token
		query = "INSERT INTO notifications_gcm (user_id, registration_token, environment) " +
			"VALUES ((SELECT user_id FROM addresses WHERE address=$1), $2, $3) "
		_, err = tx.Exec(query, address, entry.RegistrationToken, entry.Environment)
		if err != nil {
			tx.Rollback()
			return err
		}
		return tx.Commit()

	} else if existingAddress != address || existingEnv != entry.Environment {
		// update the address and environment
		query = "UPDATE notifications_gcm " +
			"SET user_id=(SELECT user_id FROM addresses WHERE address=$1), environment=$2 " +
			"WHERE registration_token=$3"
		_, err = tx.Exec(query, address, entry.Environment, entry.RegistrationToken)
		if err != nil {
			tx.Rollback()
			return err
		}
		return tx.Commit()
	}

	// row exists with the same env, nothing to do
	return tx.Rollback()
}

func (dao *NotificationsGcm) UpdateToken(address string, oldToken string, newToken string) error {
	query := "UPDATE notifications_gcm " +
		"SET registration_token=$1 " +
		"WHERE user_id=(SELECT user_id FROM addresses WHERE address=$2) " +
		"AND registration_token=$3 "
	_, err := dbconn.GetConn().Exec(query, newToken, address, oldToken)
	return err
}

func (dao *NotificationsGcm) DeleteEntry(address string, token string) (int64, error) {
	query := "DELETE FROM notifications_gcm " +
		"WHERE user_id=(SELECT user_id FROM addresses WHERE address=$1) " +
		"AND registration_token=$2 "
	result, err := dbconn.GetConn().Exec(query, address, token)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
