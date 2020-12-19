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
func Up_20130826142419(txn *sql.Tx) {
	query := `
CREATE OR REPLACE FUNCTION update_drafts_entry(IN address character varying, INOUT id integer, INOUT last_modified bigint, IN deleted boolean, IN data text, OUT conflict boolean) RETURNS record
  LANGUAGE plpgsql VOLATILE
  AS $_$
DECLARE
    original_id integer;
BEGIN
    original_id := id;
    conflict := FALSE;

    UPDATE drafts d
        SET last_modified=DEFAULT, deleted=$4, data=$5, attachments = CASE update_drafts_entry.deleted WHEN true THEN '' ELSE attachments END
        FROM users u
        WHERE d.user_id=u.id AND u.address=$1 AND d.id=$2 AND d.last_modified=$3
        RETURNING d.id, d.last_modified INTO id, last_modified;

    IF NOT FOUND THEN
        SELECT d.id, d.last_modified INTO STRICT id, last_modified
            FROM drafts d, users u
            WHERE d.user_id=u.id AND u.address=$1 AND d.id=original_id;
            conflict := TRUE;
    END IF;
END;
$_$;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20130826142419(txn *sql.Tx) {
	query := `
CREATE OR REPLACE FUNCTION update_drafts_entry(IN address character varying, INOUT id integer, INOUT last_modified bigint, IN deleted boolean, IN data text, OUT conflict boolean) RETURNS record
  LANGUAGE plpgsql VOLATILE
  AS $_$
DECLARE
    original_id integer;
BEGIN
    original_id := id;
    conflict := FALSE;

    UPDATE drafts d
        SET last_modified=DEFAULT, deleted=$4, data=$5, attachments = CASE deleted WHEN true THEN '' ELSE attachments END
        FROM users u
        WHERE d.user_id=u.id AND u.address=$1 AND d.id=$2 AND d.last_modified=$3
        RETURNING d.id, d.last_modified INTO id, last_modified;

    IF NOT FOUND THEN
        SELECT d.id, d.last_modified INTO STRICT id, last_modified
            FROM drafts d, users u
            WHERE d.user_id=u.id AND u.address=$1 AND d.id=original_id;
            conflict := TRUE;
    END IF;
END;
$_$;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
