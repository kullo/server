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
func Up_20140212172651(txn *sql.Tx) {
	query := `
DROP FUNCTION new_addrbook_id(integer);
DROP FUNCTION new_drafts_id(integer);
DROP FUNCTION new_messages_id(integer);
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20140212172651(txn *sql.Tx) {
	query := `
CREATE OR REPLACE FUNCTION new_addrbook_id(user_id integer)
  RETURNS integer AS
$BODY$
SELECT COALESCE(MAX(id), 0) + 1
FROM addrbook
WHERE user_id = $1
$BODY$
LANGUAGE sql VOLATILE;

CREATE OR REPLACE FUNCTION new_drafts_id(user_id integer)
  RETURNS integer AS
$BODY$
SELECT COALESCE(MAX(id), 0) + 1
FROM drafts
WHERE user_id = $1
$BODY$
LANGUAGE sql VOLATILE;

CREATE OR REPLACE FUNCTION new_messages_id(user_id integer)
  RETURNS integer AS
$BODY$
SELECT COALESCE(MAX(id), 0) + 1
FROM messages
WHERE user_id = $1
$BODY$
LANGUAGE sql VOLATILE;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
