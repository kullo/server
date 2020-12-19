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
func Up_20130821134324(txn *sql.Tx) {
	query := `
CREATE OR REPLACE FUNCTION delete_messages_entry(IN address character varying, INOUT id integer, INOUT last_modified bigint, OUT conflict boolean) RETURNS record
    LANGUAGE plpgsql
    AS $_$
DECLARE
    original_id integer;
BEGIN
    original_id := id;
    conflict := FALSE;

    UPDATE messages m
        SET last_modified=DEFAULT, deleted=true, received='', meta='', message='', attachments=''
        FROM users u
        WHERE m.user_id=u.id AND u.address=$1 AND m.id=$2 AND m.last_modified=$3
        RETURNING m.id, m.last_modified INTO id, last_modified;

    IF NOT FOUND THEN
        SELECT m.id, m.last_modified INTO STRICT id, last_modified
            FROM messages m, users u
            WHERE m.user_id=u.id AND u.address=$1 AND m.id=original_id;
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
func Down_20130821134324(txn *sql.Tx) {
	query := `
CREATE OR REPLACE FUNCTION delete_messages_entry(address character varying, INOUT id integer, INOUT last_modified bigint, OUT conflict boolean) RETURNS record
    LANGUAGE plpgsql
    AS $_$
DECLARE
    original_id integer;
BEGIN
    original_id := id;
    conflict := FALSE;

    UPDATE messages m
        SET last_modified=DEFAULT, deleted=true
        FROM users u
        WHERE m.user_id=u.id AND u.address=$1 AND m.id=$2 AND m.last_modified=$3
        RETURNING m.id, m.last_modified INTO id, last_modified;

    IF NOT FOUND THEN
        SELECT m.id, m.last_modified INTO STRICT id, last_modified
            FROM messages m, users u
            WHERE m.user_id=u.id AND u.address=$1 AND m.id=original_id;
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
