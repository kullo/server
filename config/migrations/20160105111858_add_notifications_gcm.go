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
func Up_20160105111858(txn *sql.Tx) {
	query := `
CREATE TABLE notifications_gcm
(
  id serial NOT NULL,
  user_id integer NOT NULL,
  registration_token character varying(254) NOT NULL,
  CONSTRAINT notifications_gcm_pkey PRIMARY KEY (id),
  CONSTRAINT notifications_gcm_unique_uid_token UNIQUE (user_id, registration_token),
  CONSTRAINT notifications_gcm_user_id_fkey FOREIGN KEY (user_id)
	REFERENCES users (id) MATCH SIMPLE
	ON UPDATE NO ACTION ON DELETE CASCADE
)
WITH (
  OIDS=FALSE
);
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20160105111858(txn *sql.Tx) {
	query := `
DROP TABLE notifications_gcm;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
