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
func Up_20150311145325(txn *sql.Tx) {
	query := `
DROP FUNCTION update_addrbook_entry(character varying, integer, bigint, boolean, text);
DROP TABLE addrbook;

DROP FUNCTION update_drafts_attachments(character varying, integer, bigint, text);
DROP FUNCTION update_drafts_entry(character varying, integer, bigint, text);
DROP TABLE drafts;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20150311145325(txn *sql.Tx) {
	query := `
CREATE TABLE drafts
(
  id integer NOT NULL,
  user_id integer NOT NULL,
  last_modified bigint NOT NULL DEFAULT kullo_now(),
  data text NOT NULL,
  attachments text NOT NULL DEFAULT ''::text,
  CONSTRAINT drafts_pkey PRIMARY KEY (id, user_id),
  CONSTRAINT drafts_user_id_fkey FOREIGN KEY (user_id)
      REFERENCES users (id) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE CASCADE
)
WITH (
  OIDS=FALSE
);

CREATE OR REPLACE FUNCTION update_drafts_entry(
    IN address character varying,
    INOUT id integer,
    INOUT last_modified bigint,
    IN data text,
    OUT conflict boolean)
  RETURNS record AS
$BODY$
DECLARE
    original_id integer;
BEGIN
    original_id := id;
    conflict := FALSE;

    UPDATE drafts d
        SET last_modified=DEFAULT, data=$4
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
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100;

CREATE OR REPLACE FUNCTION update_drafts_attachments(
    IN address character varying,
    INOUT id integer,
    INOUT last_modified bigint,
    IN attachments text,
    OUT conflict boolean)
  RETURNS record AS
$BODY$
DECLARE
    original_id integer;
BEGIN
    original_id := id;
    conflict := FALSE;

    UPDATE drafts d
        SET last_modified=DEFAULT, attachments=$4
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
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100;

CREATE TABLE addrbook
(
  id integer NOT NULL,
  user_id integer NOT NULL,
  last_modified bigint NOT NULL DEFAULT kullo_now(),
  deleted boolean NOT NULL DEFAULT false,
  data text NOT NULL,
  CONSTRAINT addrbook_pkey PRIMARY KEY (id, user_id),
  CONSTRAINT addrbook_user_id_fkey FOREIGN KEY (user_id)
      REFERENCES users (id) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE CASCADE
)
WITH (
  OIDS=FALSE
);

CREATE OR REPLACE FUNCTION update_addrbook_entry(
    IN address character varying,
    INOUT id integer,
    INOUT last_modified bigint,
    IN deleted boolean,
    IN data text,
    OUT conflict boolean)
  RETURNS record AS
$BODY$
DECLARE
    original_id integer;
BEGIN
    original_id := id;

    conflict := FALSE;
    UPDATE addrbook a
        SET last_modified=DEFAULT, deleted=$4, data=$5
        FROM users u
        WHERE a.user_id=u.id AND u.address=$1 AND a.id=$2 AND a.last_modified=$3
        RETURNING a.id, a.last_modified INTO id, last_modified;

    IF NOT FOUND THEN
        SELECT a.id, a.last_modified INTO STRICT id, last_modified
            FROM addrbook a, users u
            WHERE a.user_id=u.id AND u.address=$1 AND a.id=original_id;
            conflict := TRUE;
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
