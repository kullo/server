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
func Up_20150311165926(txn *sql.Tx) {
	query := `
CREATE TABLE addresses
(
  id serial NOT NULL,
  user_id integer NOT NULL,
  address character varying(50) NOT NULL,
  registration_code character varying(50) NOT NULL DEFAULT ''::character varying,
  CONSTRAINT addresses_pkey PRIMARY KEY (id),
  CONSTRAINT addresses_user_id_fkey FOREIGN KEY (user_id)
	REFERENCES users (id) MATCH SIMPLE
	ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT addresses_address_key UNIQUE (address)
)
WITH (
  OIDS=FALSE
);

INSERT INTO addresses (user_id, address, registration_code)
  SELECT id, address, registration_code
  FROM users;

ALTER TABLE users DROP COLUMN address;
ALTER TABLE users DROP COLUMN registration_code;


CREATE OR REPLACE FUNCTION delete_messages_entry(
    IN address character varying,
    IN id integer,
    IN last_modified bigint)
  RETURNS TABLE(id_ integer, last_modified_ bigint, conflict_ boolean) AS
$BODY$
BEGIN
	RETURN QUERY
		UPDATE messages m
		SET last_modified = DEFAULT, deleted = TRUE, received = '', meta = '', keysafe = '', content = '', attachments = NULL
		FROM addresses a
		WHERE m.user_id = a.user_id
			AND a.address = delete_messages_entry.address
			AND m.id = delete_messages_entry.id
			AND m.last_modified = delete_messages_entry.last_modified
		RETURNING m.id, m.last_modified, FALSE;

	IF NOT FOUND THEN
		RETURN QUERY
			SELECT m.id, m.last_modified, TRUE
			FROM messages m, addresses a
			WHERE m.user_id = a.user_id
				AND a.address = delete_messages_entry.address
				AND m.id = delete_messages_entry.id;
	END IF;
END;
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100
  ROWS 1000;

CREATE OR REPLACE FUNCTION update_messages_meta(
    IN address character varying,
    IN id integer,
    IN last_modified bigint,
    IN meta text)
  RETURNS TABLE(id_ integer, last_modified_ bigint, conflict_ boolean) AS
$BODY$
BEGIN
    RETURN QUERY
	UPDATE messages m
        SET last_modified = DEFAULT, meta = update_messages_meta.meta
        FROM addresses a
        WHERE m.user_id = a.user_id
		AND a.address = update_messages_meta.address
		AND m.id = update_messages_meta.id
		AND m.last_modified = update_messages_meta.last_modified
        RETURNING m.id, m.last_modified, FALSE;

    IF NOT FOUND THEN
	RETURN QUERY
	    SELECT m.id, m.last_modified, TRUE
	    FROM messages m, addresses a
	    WHERE m.user_id = a.user_id
		AND a.address = update_messages_meta.address
		AND m.id = update_messages_meta.id;
    END IF;
END;
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100
  ROWS 1000;

CREATE OR REPLACE FUNCTION upsert_keys_symm(
    address character varying,
    login_key text,
    private_data_key text)
  RETURNS void AS
$BODY$
DECLARE
	uid integer;
BEGIN
	SELECT a.user_id INTO uid
	FROM addresses a
	WHERE a.address=$1;

	UPDATE keys_symm
	SET login_key = $2, private_data_key = $3
	WHERE user_id=uid;

	IF NOT found THEN
		INSERT INTO keys_symm
		(user_id, login_key, private_data_key)
		VALUES (uid, $2, $3);
	END IF;
END;
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20150311165926(txn *sql.Tx) {
	query := `
ALTER TABLE users ADD COLUMN address character varying(50) NOT NULL DEFAULT ''::character varying;
ALTER TABLE users ADD COLUMN registration_code character varying(50) NOT NULL DEFAULT ''::character varying;

UPDATE users u
  SET
    address = a.address,
    registration_code = a.registration_code
  FROM addresses a
  WHERE u.id = a.user_id;

ALTER TABLE users
  ALTER COLUMN address DROP DEFAULT,
  ADD CONSTRAINT users_username_key UNIQUE (address);

DROP TABLE addresses;


CREATE OR REPLACE FUNCTION delete_messages_entry(
    IN address character varying,
    IN id integer,
    IN last_modified bigint)
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
  LANGUAGE plpgsql VOLATILE
  COST 100
  ROWS 1000;

CREATE OR REPLACE FUNCTION update_messages_meta(
    IN address character varying,
    IN id integer,
    IN last_modified bigint,
    IN meta text)
  RETURNS TABLE(id_ integer, last_modified_ bigint, conflict_ boolean) AS
$BODY$
BEGIN
    RETURN QUERY
	UPDATE messages m
        SET last_modified = DEFAULT, meta = update_messages_meta.meta
        FROM users u
        WHERE m.user_id = u.id
		AND u.address = update_messages_meta.address
		AND m.id = update_messages_meta.id
		AND m.last_modified = update_messages_meta.last_modified
        RETURNING m.id, m.last_modified, FALSE;

    IF NOT FOUND THEN
	RETURN QUERY
	    SELECT m.id, m.last_modified, TRUE
	    FROM messages m, users u
	    WHERE m.user_id = u.id
		AND u.address = update_messages_meta.address
		AND m.id = update_messages_meta.id;
    END IF;
END;
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100
  ROWS 1000;

CREATE OR REPLACE FUNCTION upsert_keys_symm(
    address character varying,
    login_key text,
    private_data_key text)
  RETURNS void AS
$BODY$
DECLARE
	uid integer;
BEGIN
	SELECT u.id INTO uid
	FROM users u
	WHERE u.address=$1;

	UPDATE keys_symm
	SET login_key = $2, private_data_key = $3
	WHERE user_id=uid;

	IF NOT found THEN
		INSERT INTO keys_symm
		(user_id, login_key, private_data_key)
		VALUES (uid, $2, $3);
	END IF;
END;
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
