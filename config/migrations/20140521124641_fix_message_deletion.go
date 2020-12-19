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
func Up_20140521124641(txn *sql.Tx) {
	query := `
DROP FUNCTION delete_messages_entry(character varying, integer, bigint);

CREATE OR REPLACE FUNCTION delete_messages_entry(address character varying, id integer, last_modified bigint)
  RETURNS TABLE(id_ integer, last_modified_ bigint, conflict_ boolean) AS
$BODY$
BEGIN
	RETURN QUERY
		UPDATE messages m
		SET last_modified = DEFAULT, deleted = TRUE, received = '', meta = '', keysafe = '', content = '', attachments = NULL
		FROM users u
		WHERE m.user_id = u.id
			AND u.address = delete_messages_entry.address
			AND m.id = delete_messages_entry.id
			AND m.last_modified = delete_messages_entry.last_modified
		RETURNING m.id, m.last_modified, FALSE;

	IF NOT FOUND THEN
		RETURN QUERY
			SELECT m.id, m.last_modified, TRUE
			FROM messages m, users u
			WHERE m.user_id = u.id
				AND u.address = delete_messages_entry.address
				AND m.id = delete_messages_entry.id;
	END IF;
END;
$BODY$
  LANGUAGE plpgsql VOLATILE;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20140521124641(txn *sql.Tx) {
	query := `
DROP FUNCTION delete_messages_entry(character varying, integer, bigint);

CREATE OR REPLACE FUNCTION delete_messages_entry(IN address character varying, INOUT id integer, INOUT last_modified bigint, OUT conflict boolean)
  RETURNS record AS
$BODY$
	DECLARE
		original_id integer;
	BEGIN
		original_id := id;
		conflict := FALSE;

		UPDATE messages m
			SET last_modified = DEFAULT, deleted=true, received = '', meta = '', keysafe = '', content = '', attachments = NULL
			FROM users u
			WHERE m.user_id = u.id AND u.address = delete_messages_entry.address AND m.id = delete_messages_entry.id AND m.last_modified = delete_messages_entry.last_modified
			RETURNING m.id, m.last_modified INTO id, last_modified;

		IF NOT FOUND THEN
			SELECT m.id, m.last_modified INTO STRICT id, last_modified
			    FROM messages m, users u
			    WHERE m.user_id = u.id AND u.address = delete_messages_entry.address AND m.id = original_id;
			    conflict := TRUE;
		END IF;
	END;
	$BODY$
  LANGUAGE plpgsql VOLATILE;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
