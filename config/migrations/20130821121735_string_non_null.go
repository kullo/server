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
func Up_20130821121735(txn *sql.Tx) {
	query := `
ALTER TABLE addrbook
    ALTER COLUMN data SET NOT NULL;

ALTER TABLE drafts
    ALTER COLUMN data SET NOT NULL,
    ALTER COLUMN attachments SET DEFAULT '',
    ALTER COLUMN attachments SET NOT NULL;

ALTER TABLE messages
    ALTER COLUMN received SET NOT NULL,
    ALTER COLUMN meta SET DEFAULT '',
    ALTER COLUMN meta SET NOT NULL,
    ALTER COLUMN message SET NOT NULL,
    ALTER COLUMN attachments SET DEFAULT '',
    ALTER COLUMN attachments SET NOT NULL;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20130821121735(txn *sql.Tx) {
	query := `
ALTER TABLE addrbook
    ALTER COLUMN data DROP NOT NULL;

ALTER TABLE drafts
    ALTER COLUMN data DROP NOT NULL,
    ALTER COLUMN attachments DROP NOT NULL,
    ALTER COLUMN attachments DROP DEFAULT;

ALTER TABLE messages
    ALTER COLUMN received DROP NOT NULL,
    ALTER COLUMN meta DROP NOT NULL,
    ALTER COLUMN meta DROP DEFAULT,
    ALTER COLUMN message DROP NOT NULL,
    ALTER COLUMN attachments DROP NOT NULL,
    ALTER COLUMN attachments DROP DEFAULT;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
