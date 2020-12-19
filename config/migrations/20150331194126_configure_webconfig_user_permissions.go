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
func Up_20150331194126(txn *sql.Tx) {
	query := `
GRANT SELECT ON ALL TABLES IN SCHEMA public TO webconfig;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO webconfig;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Down is executed when this migration is rolled back
func Down_20150331194126(txn *sql.Tx) {
	query := `
ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE SELECT ON TABLES FROM webconfig;
REVOKE SELECT ON ALL TABLES IN SCHEMA public FROM webconfig;
`
	_, err := txn.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
