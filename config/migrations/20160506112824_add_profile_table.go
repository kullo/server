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
func Up_20160506112824(txn *sql.Tx) {
	query := `
CREATE TABLE profile
(
  user_id integer NOT NULL,
  key character varying(64) NOT NULL,
  value text,
  last_modified bigint DEFAULT kullo_now() NOT NULL,
  CONSTRAINT profile_user_id_fkey FOREIGN KEY (user_id)
	REFERENCES users (id) MATCH SIMPLE
	ON UPDATE NO ACTION ON DELETE CASCADE,
  PRIMARY KEY (user_id, key)
);
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20160506112824(txn *sql.Tx) {
	query := `
DROP TABLE profile;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
