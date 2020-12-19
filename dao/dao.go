/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package dao

import (
	"database/sql"
	"errors"
)

var ErrConflict = errors.New("dao: conflicting modification")

type ID struct {
	ID uint32 `json:"id"`
}

type IDLastModified struct {
	ID           uint32 `json:"id"`
	LastModified uint64 `json:"lastModified"` // timestamp*10^6, µs since the epoch
}

func GetNextIDLastModified(rows *sql.Rows) (*IDLastModified, error) {
	idlc := &IDLastModified{}
	err := rows.Scan(&idlc.ID, &idlc.LastModified)
	return idlc, err
}
