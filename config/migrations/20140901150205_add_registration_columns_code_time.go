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
func Up_20140901150205(txn *sql.Tx) {
	query := `
		ALTER TABLE users ADD COLUMN registration_time timestamp with time zone default NOW();
		ALTER TABLE users ALTER COLUMN registration_time SET NOT NULL;

		ALTER TABLE users ADD COLUMN registration_code character varying(50) default '';
		ALTER TABLE users ALTER COLUMN registration_code SET NOT NULL;
		`

	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20140901150205(txn *sql.Tx) {
	query := `
	    ALTER TABLE users DROP COLUMN registration_time;
	    ALTER TABLE users DROP COLUMN registration_code;
	    `

	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
