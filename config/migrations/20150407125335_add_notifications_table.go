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
func Up_20150407125335(txn *sql.Tx) {
	query := `
CREATE TABLE notifications
(
  id serial NOT NULL,
  user_id integer NOT NULL,
  email character varying(254) NOT NULL,
  double_opt_in_secret character varying(16) NOT NULL DEFAULT kullo_random_id(),
  CONSTRAINT notifications_pkey PRIMARY KEY (id),
  CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id)
	REFERENCES users (id) MATCH SIMPLE
	ON UPDATE NO ACTION ON DELETE CASCADE
)
WITH (
  OIDS=FALSE
);

GRANT INSERT, UPDATE, DELETE ON notifications TO webconfig;
GRANT USAGE ON SEQUENCE notifications_id_seq TO webconfig;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20150407125335(txn *sql.Tx) {
	query := `
REVOKE INSERT, UPDATE, DELETE ON notifications FROM webconfig;
REVOKE USAGE ON SEQUENCE notifications_id_seq FROM webconfig;

DROP TABLE notifications;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
