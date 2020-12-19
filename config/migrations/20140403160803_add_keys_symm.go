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
func Up_20140403160803(txn *sql.Tx) {
	query := `
CREATE TABLE keys_symm
(
  user_id integer NOT NULL,
  login_key text NOT NULL,
  private_data_key text NOT NULL,
  CONSTRAINT keys_symm_pkey PRIMARY KEY (user_id),
  CONSTRAINT keys_asymm_user_id_fkey FOREIGN KEY (user_id)
      REFERENCES users (id) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE CASCADE
)
WITH (
  OIDS=FALSE
);

CREATE OR REPLACE FUNCTION upsert_keys_symm(address character varying, login_key text, private_data_key text)
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
  LANGUAGE plpgsql VOLATILE;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20140403160803(txn *sql.Tx) {
	query := `
DROP TABLE keys_symm;
DROP FUNCTION upsert_keys_symm(character varying, text, text);
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
