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
func Up_20170217145502(txn *sql.Tx) {
	query := `
ALTER TABLE users
	ADD COLUMN language text NOT NULL DEFAULT ''::text;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20170217145502(txn *sql.Tx) {
	query := `
ALTER TABLE users
	DROP COLUMN language;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
