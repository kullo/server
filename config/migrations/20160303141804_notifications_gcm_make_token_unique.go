/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package main

import (
	"database/sql"
	"log"
)

// Up is executed when this migration is applied
func Up_20160303141804(txn *sql.Tx) {
	query := `
ALTER TABLE notifications_gcm
	DROP CONSTRAINT notifications_gcm_unique_uid_token,
	ADD CONSTRAINT notifications_gcm_unique_token UNIQUE (registration_token);
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20160303141804(txn *sql.Tx) {
	query := `
ALTER TABLE notifications_gcm
	DROP CONSTRAINT notifications_gcm_unique_token,
	ADD CONSTRAINT notifications_gcm_unique_uid_token UNIQUE (user_id, registration_token);
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
