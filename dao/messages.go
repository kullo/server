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

	"bitbucket.org/kullo/server/dbconn"
)

const KIBIBYTE int = 1024
const MEBIBYTE int = 1024 * KIBIBYTE

const MESSAGE_KEY_SAFE_MAX_BYTES int = 1 * KIBIBYTE
const MESSAGE_CONTENT_MAX_BYTES int = 128 * KIBIBYTE
const MESSAGE_META_MAX_BYTES int = 1 * KIBIBYTE
const MESSAGE_JSON_ATTACHMENTS_MAX_BYTES int = 16 * MEBIBYTE
const MESSAGE_ATTACHMENTS_MAX_BYTES int = 100 * MEBIBYTE

type MessagesEntry struct {
	ID                uint32 `json:"id"`
	LastModified      uint64 `json:"lastModified"` // timestamp*10^6, µs since the epoch
	Deleted           bool   `json:"deleted"`
	Received          string `json:"dateReceived"`
	Meta              string `json:"meta"`
	KeySafe           string `json:"keySafe"`
	Content           string `json:"content"`
	HasAttachments    bool   `json:"hasAttachments"`
	AttachmentsBase64 string `json:"attachments,omitempty"`
	Attachments       []byte `json:"-"`
}

func (e *MessagesEntry) SetIDLastModifiedDeleted(id uint32, lastModified uint64, deleted bool) {
	e.ID = id
	e.LastModified = lastModified
	e.Deleted = deleted
}

func (e *MessagesEntry) ValidForCreation() bool {
	// base64 string length in bytes * 3/4 = binary length
	return (e.KeySafe != "" && len(e.KeySafe)*3 <= MESSAGE_KEY_SAFE_MAX_BYTES*4) &&
		(e.Content != "" && len(e.Content)*3 <= MESSAGE_CONTENT_MAX_BYTES*4) &&
		(len(e.Meta)*3 <= MESSAGE_META_MAX_BYTES*4) &&
		(len(e.Attachments) <= MESSAGE_ATTACHMENTS_MAX_BYTES)
}

func (e *MessagesEntry) ValidForModification() bool {
	return len(e.Meta)*3 <= MESSAGE_META_MAX_BYTES*4
}

type Messages struct {
}

func (dao *Messages) GetList(address string, modifiedAfter uint64, includeData bool) (uint32, uint32, *sql.Rows, error) {
	rowsCount, err := dbconn.GetConn().Query("SELECT COUNT(m.id) "+
		"FROM messages m JOIN addresses a USING (user_id) "+
		"WHERE a.address=$1 AND m.last_modified > $2",
		address, modifiedAfter)
	if err != nil {
		return 0, 0, nil, err
	}
	defer rowsCount.Close()

	if !rowsCount.Next() {
		return 0, 0, nil, errors.New("Couldn't get messages count")
	}
	var resultsTotal uint32
	err = rowsCount.Scan(&resultsTotal)
	if err != nil {
		return 0, 0, nil, err
	}
	var resultsReturned uint32 = 100
	if resultsReturned > resultsTotal {
		resultsReturned = resultsTotal
	}

	fields := "m.id, m.last_modified"
	if includeData {
		fields += ", m.deleted, m.received, m.meta, m.keysafe, m.content, " +
			"m.attachments IS NOT NULL"
	}
	rows, err := dbconn.GetConn().
		Query("SELECT "+fields+" "+
			"FROM messages m JOIN addresses a USING (user_id) "+
			"WHERE a.address=$1 AND m.last_modified > $2 "+
			"ORDER BY m.last_modified ASC "+
			"LIMIT $3",
			address, modifiedAfter, resultsReturned)
	return resultsTotal, resultsReturned, rows, err
}

func (dao *Messages) GetUnreadCount(address string) uint32 {
	var count uint32
	dbconn.GetConn().QueryRow("SELECT count(*) FROM messages "+
		"WHERE user_id = (SELECT user_id FROM addresses WHERE address = $1) "+
		"AND meta = '' AND deleted = false AND received >= '2016-01-01T00:00:00Z'",
		address).Scan(&count)
	return count
}

func (dao *Messages) GetStorageSize(address string) (uint64, error) {
	var storageSize uint64
	err := dbconn.GetConn().QueryRow(
		"SELECT coalesce(sum( "+
			"coalesce(octet_length(content), 0) + "+
			"coalesce(octet_length(keysafe), 0) + "+
			"coalesce(octet_length(attachments), 0) "+
			"), 0) "+
			"FROM messages "+
			"WHERE user_id = (SELECT user_id FROM addresses WHERE address = $1) ",
		address).Scan(&storageSize)
	return storageSize, err
}

func (dao *Messages) GetNextEntry(rows *sql.Rows) (*MessagesEntry, error) {
	entry := &MessagesEntry{}
	err := rows.Scan(&entry.ID, &entry.LastModified, &entry.Deleted, &entry.Received, &entry.Meta, &entry.KeySafe, &entry.Content, &entry.HasAttachments)
	return entry, err
}

func (dao *Messages) InsertEntry(address string, entry *MessagesEntry) error {
	query := "WITH usr AS (SELECT user_id FROM addresses WHERE address=$1) " +
		"INSERT INTO messages (id, user_id, received, keysafe, content, attachments, meta) " +
		"VALUES (kullo_new_id('messages', (SELECT user_id FROM usr)), " +
		"(SELECT user_id FROM usr), $2, $3, $4, $5, $6) " +
		"RETURNING id, last_modified"
	var att *[]byte
	if len(entry.Attachments) > 0 {
		att = &entry.Attachments
	}
	err := dbconn.GetConn().
		QueryRow(query, address, entry.Received, entry.KeySafe, entry.Content, att, entry.Meta).
		Scan(&entry.ID, &entry.LastModified)
	return err
}

func (dao *Messages) GetEntry(address string, id uint32) (*MessagesEntry, error) {
	entry := &MessagesEntry{}
	err := dbconn.GetConn().
		QueryRow("SELECT m.id, m.last_modified, m.deleted, m.received, m.meta, "+
			"m.keysafe, m.content, m.attachments IS NOT NULL "+
			"FROM messages m JOIN addresses a USING (user_id) "+
			"WHERE a.address=$1 AND m.id=$2", address, id).
		Scan(&entry.ID, &entry.LastModified, &entry.Deleted, &entry.Received, &entry.Meta, &entry.KeySafe, &entry.Content, &entry.HasAttachments)
	return entry, err
}

func (dao *Messages) ModifyMeta(address string, entry *MessagesEntry) (*IDLastModified, error) {
	var meta IDLastModified
	var conflict bool
	err := dbconn.GetConn().
		QueryRow("SELECT id_, last_modified_, conflict_ "+
			"FROM update_messages_meta($1, $2, $3, $4)",
			address, entry.ID, entry.LastModified, entry.Meta).
		Scan(&meta.ID, &meta.LastModified, &conflict)
	if conflict {
		return &meta, ErrConflict
	}
	return &meta, err
}

func (dao *Messages) DeleteEntry(address string, id uint32, lastModified uint64) (*IDLastModified, error) {
	var meta IDLastModified
	var conflict bool
	err := dbconn.GetConn().
		QueryRow("SELECT id_, last_modified_, conflict_ "+
			"FROM delete_messages_entry($1, $2, $3)",
			address, id, lastModified).
		Scan(&meta.ID, &meta.LastModified, &conflict)
	if conflict {
		return &meta, ErrConflict
	}
	return &meta, err
}

func (dao *Messages) GetAttachments(address string, id uint32) ([]byte, error) {
	var attachments []byte
	err := dbconn.GetConn().
		QueryRow("SELECT m.attachments "+
			"FROM messages m JOIN addresses a USING (user_id) "+
			"WHERE a.address=$1 AND m.id=$2 AND m.attachments IS NOT NULL",
			address, id).
		Scan(&attachments)
	return attachments, err
}
