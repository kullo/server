/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package main

import (
	"database/sql"
	"encoding/base64"
	"log"
)

// Up is executed when this migration is applied
func Up_20140513150851(txn *sql.Tx) {
	query := `
	ALTER TABLE messages ADD COLUMN attachments_blob bytea;
	`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}

	for {
		var id uint32
		var userId uint32
		var attachments string

		query = `
		SELECT id, user_id, attachments FROM messages WHERE attachments != '' LIMIT 1;
		`
		err = txn.QueryRow(query).Scan(&id, &userId, &attachments)
		if err == sql.ErrNoRows {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		attachmentsBase64, err := base64.StdEncoding.DecodeString(attachments)
		if err != nil {
			log.Fatal(err)
		}
		query = `
		UPDATE messages
			SET attachments = '', attachments_blob = $1
			WHERE id = $2 AND user_id = $3;
		`
		_, err = txn.Exec(query, attachmentsBase64, id, userId)
		if err != nil {
			log.Fatal(err)
		}
	}

	query = `
	ALTER TABLE messages DROP COLUMN attachments;
	ALTER TABLE messages RENAME COLUMN attachments_blob TO attachments;

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
	_, err = txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20140513150851(txn *sql.Tx) {
	query := `
	ALTER TABLE messages ADD COLUMN attachments_text text NOT NULL DEFAULT ''::text;
	`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}

	for {
		var id uint32
		var userId uint32
		var attachments []byte

		query = `
		SELECT id, user_id, attachments FROM messages WHERE attachments IS NOT NULL LIMIT 1;
		`
		err = txn.QueryRow(query).Scan(&id, &userId, &attachments)
		if err == sql.ErrNoRows {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		attachmentsText := base64.StdEncoding.EncodeToString(attachments)
		if err != nil {
			log.Fatal(err)
		}
		query = `
		UPDATE messages
			SET attachments = NULL, attachments_text = $1
			WHERE id = $2 AND user_id = $3;
		`
		_, err = txn.Exec(query, attachmentsText, id, userId)
		if err != nil {
			log.Fatal(err)
		}
	}

	query = `
	ALTER TABLE messages DROP COLUMN attachments;
	ALTER TABLE messages RENAME COLUMN attachments_text TO attachments;

	CREATE OR REPLACE FUNCTION delete_messages_entry(IN address character varying, INOUT id integer, INOUT last_modified bigint, OUT conflict boolean)
	  RETURNS record AS
	$BODY$
	DECLARE
		original_id integer;
	BEGIN
		original_id := id;
		conflict := FALSE;

		UPDATE messages m
		    SET last_modified=DEFAULT, deleted=true, received='', meta='', keysafe='', content='', attachments=NULL
		    FROM users u
		    WHERE m.user_id=u.id AND u.address=address AND m.id=id AND m.last_modified=last_modified
		    RETURNING m.id, m.last_modified INTO id, last_modified;

		IF NOT FOUND THEN
		    SELECT m.id, m.last_modified INTO STRICT id, last_modified
		        FROM messages m, users u
		        WHERE m.user_id=u.id AND u.address=address AND m.id=original_id;
		        conflict := TRUE;
		END IF;
	END;
	$BODY$
	  LANGUAGE plpgsql VOLATILE;
	`
	_, err = txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
