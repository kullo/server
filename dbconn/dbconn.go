/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package dbconn

import (
	"database/sql"
)

var db *sql.DB

func Open(driverName string, dataSourceName string) error {
	var err error
	db, err = sql.Open(driverName, dataSourceName)
	if err == nil {
		db.SetMaxOpenConns(90)
		db.SetMaxIdleConns(10)
	}
	return err
}

func Close() {
	db.Close()
}

func GetConn() *sql.DB {
	return db
}
