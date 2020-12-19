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
func Up_20170529151643(txn *sql.Tx) {
	query := `
CREATE INDEX keys_asymm__user_id
  ON keys_asymm
  (user_id);

CREATE INDEX messages__user_id
  ON messages
  (user_id);

CREATE INDEX messages__user_id__last_modified
  ON messages
  (user_id, last_modified);

CREATE INDEX messages__user_id__received__unread
  ON messages
  (user_id, received)
  WHERE meta = ''::text AND deleted = false;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20170529151643(txn *sql.Tx) {
	query := `
DROP INDEX messages__user_id__received__unread;
DROP INDEX messages__user_id__last_modified;
DROP INDEX messages__user_id;
DROP INDEX keys_asymm__user_id;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
