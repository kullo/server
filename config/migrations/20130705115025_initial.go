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
func Up_20130705115025(txn *sql.Tx) {
	query := `
SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

SET search_path = public, pg_catalog;


CREATE FUNCTION delete_messages_entry(address character varying, INOUT id integer, INOUT last_modified bigint, OUT conflict boolean) RETURNS record
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


CREATE FUNCTION kullo_now() RETURNS bigint
    LANGUAGE sql
    AS $$SELECT (EXTRACT(epoch FROM now())*1000000)::bigint$$;


CREATE FUNCTION new_addrbook_id(user_id integer) RETURNS integer
    LANGUAGE sql
    AS $_$SELECT COALESCE(MAX(id), 0) + 1
FROM addrbook
WHERE user_id = $1$_$;


CREATE FUNCTION new_drafts_id(user_id integer) RETURNS integer
    LANGUAGE sql
    AS $_$SELECT COALESCE(MAX(id), 0) + 1
FROM drafts
WHERE user_id = $1$_$;


CREATE FUNCTION new_messages_id(user_id integer) RETURNS integer
    LANGUAGE sql
    AS $_$SELECT COALESCE(MAX(id), 0) + 1
FROM messages
WHERE user_id = $1$_$;


CREATE FUNCTION update_addrbook_entry(address character varying, INOUT id integer, INOUT last_modified bigint, deleted boolean, data text, OUT conflict boolean) RETURNS record
    LANGUAGE plpgsql
    AS $_$
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
$_$;


CREATE FUNCTION update_drafts_attachments(address character varying, INOUT id integer, INOUT last_modified bigint, attachments text, OUT conflict boolean) RETURNS record
    LANGUAGE plpgsql
    AS $_$
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
$_$;


CREATE FUNCTION update_drafts_entry(address character varying, INOUT id integer, INOUT last_modified bigint, deleted boolean, data text, OUT conflict boolean) RETURNS record
    LANGUAGE plpgsql
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


CREATE FUNCTION update_messages_meta(address character varying, INOUT id integer, INOUT last_modified bigint, meta text, OUT conflict boolean) RETURNS record
    LANGUAGE plpgsql
    AS $_$
DECLARE
    original_id integer;
BEGIN
    original_id := id;
    conflict := FALSE;

    UPDATE messages m
        SET last_modified=DEFAULT, meta=$4
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


SET default_tablespace = '';

SET default_with_oids = false;


CREATE TABLE addrbook (
    id integer NOT NULL,
    user_id integer NOT NULL,
    last_modified bigint DEFAULT kullo_now() NOT NULL,
    deleted boolean DEFAULT false NOT NULL,
    data text
);


CREATE TABLE drafts (
    id integer NOT NULL,
    user_id integer NOT NULL,
    last_modified bigint DEFAULT kullo_now() NOT NULL,
    deleted boolean DEFAULT false NOT NULL,
    data text,
    attachments text
);


CREATE TABLE messages (
    id integer NOT NULL,
    user_id integer NOT NULL,
    last_modified bigint DEFAULT kullo_now() NOT NULL,
    deleted boolean DEFAULT false NOT NULL,
    received character varying(32),
    meta text,
    message text,
    attachments text
);


CREATE TABLE users (
    id integer NOT NULL,
    address character varying(50) NOT NULL
);


CREATE SEQUENCE users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE users_id_seq OWNED BY users.id;


ALTER TABLE ONLY users ALTER COLUMN id SET DEFAULT nextval('users_id_seq'::regclass);


ALTER TABLE ONLY addrbook
    ADD CONSTRAINT addrbook_pkey PRIMARY KEY (id, user_id);


ALTER TABLE ONLY drafts
    ADD CONSTRAINT drafts_pkey PRIMARY KEY (id, user_id);


ALTER TABLE ONLY messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id, user_id);


ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


ALTER TABLE ONLY users
    ADD CONSTRAINT users_username_key UNIQUE (address);


ALTER TABLE ONLY addrbook
    ADD CONSTRAINT addrbook_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;


ALTER TABLE ONLY drafts
    ADD CONSTRAINT drafts_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;


ALTER TABLE ONLY messages
    ADD CONSTRAINT messages_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20130705115025(txn *sql.Tx) {
	log.Fatal("Undoing the initial migration is not supported.")
}
