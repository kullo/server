/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package dao

import (
	"database/sql"

	"bitbucket.org/kullo/server/dbconn"
)

var validProfileKeys []string = []string{
	"name", "organization", "footer", "avatar_type", "avatar_data", "mk_backup_reminder",
}

type ProfileKeyLastModified struct {
	Key          string `json:"key"`
	LastModified uint64 `json:"lastModified"` // timestamp*10^6, µs since the epoch
}

func (entry *ProfileKeyLastModified) IsValid() bool {
	found := false
	for _, key := range validProfileKeys {
		if key == entry.Key {
			found = true
			break
		}
	}
	return found
}

type ProfileEntry struct {
	ProfileKeyLastModified
	Value string `json:"value"`
}

func (entry *ProfileEntry) SetKeyLastModified(key string, lastModified uint64) {
	entry.Key = key
	entry.LastModified = lastModified
}

type Profile struct {
}

func (dao *Profile) GetList(address string, modifiedAfter uint64) (*sql.Rows, error) {
	rows, err := dbconn.GetConn().
		Query("SELECT p.key, p.value, p.last_modified "+
			"FROM profile p JOIN addresses a USING (user_id) "+
			"WHERE a.address=$1 AND p.last_modified > $2 "+
			"ORDER BY p.last_modified ASC ",
			address, modifiedAfter)
	return rows, err
}

func (dao *Profile) GetNextEntry(rows *sql.Rows) (*ProfileEntry, error) {
	entry := &ProfileEntry{}
	err := rows.Scan(&entry.Key, &entry.Value, &entry.LastModified)
	return entry, err
}

func (dao *Profile) GetEntry(address string, key string) (*ProfileEntry, error) {
	entry := &ProfileEntry{}
	err := dbconn.GetConn().
		QueryRow("SELECT p.key, p.value, p.last_modified "+
			"FROM profile p JOIN addresses a USING (user_id) "+
			"WHERE a.address=$1 AND p.key=$2", address, key).
		Scan(&entry.Key, &entry.Value, &entry.LastModified)
	return entry, err
}

func (dao *Profile) insertEntry(tx *sql.Tx, address string, entry *ProfileEntry) (*ProfileKeyLastModified, error) {
	query := "WITH usr AS (SELECT user_id FROM addresses WHERE address=$1) " +
		"INSERT INTO profile (user_id, key, value) " +
		"VALUES ((SELECT user_id FROM usr), $2, $3) " +
		"RETURNING last_modified"
	result := ProfileKeyLastModified{Key: entry.Key}
	err := tx.
		QueryRow(query, address, entry.Key, entry.Value).
		Scan(&result.LastModified)
	return &result, err
}

func (dao *Profile) ModifyEntry(address string, entry *ProfileEntry) (*ProfileKeyLastModified, error) {
	tx, err := dbconn.GetConn().Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// check last
	var lastModified uint64
	err = tx.QueryRow("SELECT p.last_modified "+
		"FROM profile p JOIN addresses a USING (user_id) "+
		"WHERE a.address=$1 AND p.key=$2",
		address, entry.Key).Scan(&lastModified)
	if err == sql.ErrNoRows {
		if entry.LastModified != 0 {
			// trying to update an entry that the server doesn't have
			return nil, err
		}
		meta, err := dao.insertEntry(tx, address, entry)
		if err == nil {
			tx.Commit()
		}
		return meta, err
	}
	if err != nil {
		return nil, err
	}
	if entry.LastModified != lastModified {
		return nil, ErrConflict
	}

	var meta ProfileKeyLastModified
	err = tx.QueryRow("WITH usr AS (SELECT user_id FROM addresses WHERE address=$1) "+
		"UPDATE profile "+
		"SET value = $2, last_modified = DEFAULT "+
		"WHERE user_id=(SELECT user_id FROM usr) AND key=$3 AND last_modified=$4 "+
		"RETURNING key, last_modified ",
		address, entry.Value, entry.Key, entry.LastModified).
		Scan(&meta.Key, &meta.LastModified)
	if err == nil {
		tx.Commit()
	}
	return &meta, err
}
