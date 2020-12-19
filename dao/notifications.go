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

type NotificationsEntry struct {
	Email            string
	WebloginUsername string
	CancelSecret     string
}

type Notifications struct {
}

func (dao *Notifications) GetConfirmedEntry(address string) (*NotificationsEntry, error) {
	var entry NotificationsEntry
	err := dbconn.GetConn().
		QueryRow("SELECT n.email, u.weblogin_username, n.double_opt_in_secret "+
			"FROM notifications n "+
			"JOIN addresses a USING (user_id) "+
			"JOIN users u ON n.user_id = u.id "+
			"WHERE a.address=$1 AND n.confirmed IS NOT NULL ",
			address).
		Scan(&entry.Email, &entry.WebloginUsername, &entry.CancelSecret)
	return &entry, err
}
