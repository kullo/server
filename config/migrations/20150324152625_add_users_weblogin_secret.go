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
func Up_20150324152625(txn *sql.Tx) {
	query := `
CREATE OR REPLACE FUNCTION kullo_random_id()
  RETURNS text AS
$BODY$
SELECT substr(translate( -- truncate after replacing + by - and / by _
	encode(decode( -- transcode hex -> base64
		md5( -- each call to random() yields ~47b of randomness
			to_char(random(), '0.99999999999999') ||
			to_char(random(), '0.99999999999999')
		),
		'hex'), 'base64'),
	'+/', '-_'), 1, 16) -- 16 base64 chars = 12 bytes = 96b
$BODY$
  LANGUAGE sql VOLATILE
  COST 100;

ALTER TABLE users ADD COLUMN weblogin_username varchar(16) NOT NULL DEFAULT kullo_random_id() UNIQUE;
ALTER TABLE users ADD COLUMN weblogin_secret varchar(16) NOT NULL DEFAULT kullo_random_id();
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20150324152625(txn *sql.Tx) {
	query := `
ALTER TABLE users DROP COLUMN weblogin_secret;
ALTER TABLE users DROP COLUMN weblogin_username;

DROP FUNCTION kullo_random_id();
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
