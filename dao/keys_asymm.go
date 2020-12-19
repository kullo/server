/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package dao

import (
	"database/sql"
	"math"
	"time"

	"bitbucket.org/kullo/server/dbconn"
)

type PubkeyEntry struct {
	ID         uint32    `json:"id"`
	Type       string    `json:"type"`
	Pubkey     string    `json:"pubkey"`
	ValidFrom  time.Time `json:"validFrom"`
	ValidUntil time.Time `json:"validUntil"`
	Revocation string    `json:"revocation"`
}

type KeysAsymmEntry struct {
	PubkeyEntry
	Privkey string `json:"privkey"`
}

type KeysAsymm struct {
}

const LATEST_ENCRYPTION_PUBKEY = math.MaxUint32

func (dao *KeysAsymm) GetList(address string, keyType string, date time.Time) (*sql.Rows, error) {
	fields := "ka.id, ka.key_type, ka.pubkey, ka.privkey, ka.valid_from, ka.valid_until, ka.revocation"
	whereExtra := ""
	var argsExtra []interface{}
	if keyType != "" {
		whereExtra += " AND key_type=$2"
		argsExtra = append(argsExtra, keyType)
	}
	if !date.IsZero() {
		//TODO implement filtering by date
	}

	var args []interface{}
	args = append(args, address)
	return dbconn.GetConn().
		Query("SELECT "+fields+" "+
			"FROM keys_asymm ka JOIN addresses a ON ka.user_id=a.user_id "+
			"WHERE a.address=$1"+whereExtra+" "+
			"ORDER BY ka.id ASC",
			append(args, argsExtra...)...)
}

func (dao *KeysAsymm) GetNextEntry(rows *sql.Rows) (*KeysAsymmEntry, error) {
	entry := &KeysAsymmEntry{}
	err := rows.Scan(
		&entry.ID, &entry.Type, &entry.Pubkey, &entry.Privkey,
		&entry.ValidFrom, &entry.ValidUntil, &entry.Revocation)
	return entry, err
}

func (dao *KeysAsymm) InsertEntry(address string, entry *KeysAsymmEntry) (*ID, error) {
	//TODO replace this with a stored procedure that can react on the WITH select returning nothing (also at other occurrences of WITH)
	var id ID
	err := dbconn.GetConn().
		QueryRow(
			"WITH addr AS (SELECT user_id FROM addresses WHERE address=$1) "+
				"INSERT INTO keys_asymm "+
				"(id, user_id, key_type, pubkey, privkey, valid_from, valid_until) "+
				"VALUES "+
				"(kullo_new_id('keys_asymm', (SELECT user_id FROM addr)), "+
				"(SELECT user_id FROM addr), $2, $3, $4, $5, $6) "+
				"RETURNING id",
			address, entry.Type, entry.Pubkey, entry.Privkey, entry.ValidFrom, entry.ValidUntil).
		Scan(&id.ID)
	return &id, err
}

func (dao *KeysAsymm) GetEntry(address string, id uint32) (*KeysAsymmEntry, error) {
	entry := &KeysAsymmEntry{}

	var row *sql.Row
	if id == LATEST_ENCRYPTION_PUBKEY {
		row = dbconn.GetConn().
			QueryRow("SELECT ka.id, ka.key_type, ka.pubkey, ka.privkey, "+
				"ka.valid_from, ka.valid_until, ka.revocation "+
				"FROM keys_asymm ka JOIN addresses a ON ka.user_id=a.user_id "+
				"WHERE a.address=$1 AND ka.key_type='enc' "+
				"ORDER BY ka.valid_from DESC LIMIT 1", address)
	} else {
		row = dbconn.GetConn().
			QueryRow("SELECT ka.id, ka.key_type, ka.pubkey, ka.privkey, "+
				"ka.valid_from, ka.valid_until, ka.revocation "+
				"FROM keys_asymm ka JOIN addresses a ON ka.user_id=a.user_id "+
				"WHERE a.address=$1 AND ka.id=$2", address, id)
	}
	err := row.Scan(
		&entry.ID, &entry.Type, &entry.Pubkey, &entry.Privkey,
		&entry.ValidFrom, &entry.ValidUntil, &entry.Revocation)

	return entry, err
}

func (dao *KeysAsymm) SetRevocation(address string, id uint32, revocation string) error {
	_, err := dbconn.GetConn().
		Exec("UPDATE keys_asymm ka SET revocation=$1 "+
			"WHERE ka.id=$2 AND ka.user_id=(SELECT user_id FROM addresses WHERE address=$3)",
			revocation, id, address)
	return err
}
