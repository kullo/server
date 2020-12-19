/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package dao

import (
	"bitbucket.org/kullo/server/dbconn"
)

type KeysSymmEntry struct {
	LoginKey       string `json:"loginKey,omitempty"`
	PrivateDataKey string `json:"privateDataKey"`
}

type KeysSymm struct {
}

func (dao *KeysSymm) InsertOrUpdateEntry(address string, entry *KeysSymmEntry) error {
	_, err := dbconn.GetConn().
		Exec("SELECT upsert_keys_symm($1, $2, $3)",
			address, entry.LoginKey, entry.PrivateDataKey)
	return err
}

func (dao *KeysSymm) GetEntry(address string) (*KeysSymmEntry, error) {
	entry := &KeysSymmEntry{}
	err := dbconn.GetConn().
		QueryRow("SELECT ks.login_key, ks.private_data_key "+
			"FROM keys_symm ks JOIN addresses a ON ks.user_id=a.user_id "+
			"WHERE a.address=$1", address).
		Scan(&entry.LoginKey, &entry.PrivateDataKey)
	return entry, err
}
