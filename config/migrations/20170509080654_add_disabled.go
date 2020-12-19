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
func Up_20170509080654(txn *sql.Tx) {
	query := `
ALTER TABLE users
	ADD COLUMN disabled boolean NOT NULL DEFAULT FALSE;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20170509080654(txn *sql.Tx) {
	query := `
ALTER TABLE users
	DROP COLUMN disabled;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
