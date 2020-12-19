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
func Up_20170127102611(txn *sql.Tx) {
	query := `
CREATE TABLE plans
(
  id serial NOT NULL,
  name character varying(50) NOT NULL,
  storage_quota bigint NOT NULL,
  CONSTRAINT plans_pkey PRIMARY KEY (id),
  CONSTRAINT plans_name_key UNIQUE (name)
)
WITH (
  OIDS=FALSE
);

INSERT INTO plans (name, storage_quota) VALUES
	('Free', 1 * cast(1024*1024*1024 as bigint)),
	('Friend', 10 * cast(1024*1024*1024 as bigint)),
	('Professional', 15 * cast(1024*1024*1024 as bigint));

ALTER TABLE users
	ADD COLUMN plan_id integer,
	ADD CONSTRAINT users_plan_id_fkey FOREIGN KEY (plan_id)
		REFERENCES plans (id) MATCH SIMPLE
		ON UPDATE NO ACTION ON DELETE RESTRICT;

UPDATE users u
	SET plan_id = (SELECT id FROM plans WHERE name = 'Professional')
	FROM addresses a
	WHERE u.id = a.user_id AND
		right(a.address, 10) != '#kullo.net' AND
		right(a.address, 11) != '#kullo.test';

UPDATE users u
	SET plan_id = (SELECT id FROM plans WHERE name = 'Free')
	WHERE plan_id IS NULL;

ALTER TABLE users ALTER COLUMN plan_id SET NOT NULL;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20170127102611(txn *sql.Tx) {
	query := `
ALTER TABLE users
	DROP CONSTRAINT users_plan_id_fkey,
	DROP COLUMN plan_id;

DROP TABLE plans;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
