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
func Up_20140212124908(txn *sql.Tx) {
	query := `
CREATE OR REPLACE FUNCTION kullo_new_id(the_table regclass, user_id integer, OUT result integer)
AS
$BODY$
BEGIN
    EXECUTE format('
	    SELECT COALESCE(MAX(id), 0) + 1
	    FROM %s
	    WHERE user_id = %s
    ', the_table, user_id) INTO result;
END
$BODY$
LANGUAGE plpgsql VOLATILE;

CREATE TABLE keys_asymm
(
	id integer NOT NULL,
	user_id integer NOT NULL,
	key_type character varying(3) NOT NULL,
	pubkey text NOT NULL,
	privkey text NOT NULL,
	valid_from timestamp with time zone NOT NULL,
	valid_until timestamp with time zone NOT NULL,
	revocation text NOT NULL DEFAULT '',
	CONSTRAINT keys_asymm_pkey PRIMARY KEY (id, user_id),
	CONSTRAINT keys_asymm_user_id_fkey FOREIGN KEY (user_id)
		REFERENCES users (id) MATCH SIMPLE
		ON UPDATE NO ACTION ON DELETE CASCADE
);
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20140212124908(txn *sql.Tx) {
	query := `
DROP FUNCTION kullo_new_id(regclass, integer);
DROP TABLE keys_asymm;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
