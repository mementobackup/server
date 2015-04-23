/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package database

import (
	"bitbucket.org/ebianchi/memento-common/common"
	"database/sql"
	"github.com/op/go-logging"
)

func Getdataset(db *DB, grace string) int {
	var result int

	var query = "SELECT actual FROM status WHERE grace = ?"

	err := db.Conn.QueryRow(query, grace).Scan(&result)
	if err != nil {
		// TODO: add log point
		return 0
	}
	return result
}

func Saveattrs(log *logging.Logger, db *DB, section *common.Section, metadata common.JSONFile) {
	var tx *sql.Tx
	var stmt *sql.Stmt
	var err error

	var insert = "INSERT INTO attrs" +
		"(area, grace, dataset, element, os, username, groupname, type," +
		" link, mtime, ctime, hash, perms, compressed)" +
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

	tx, err = db.Conn.Begin()
	if err != nil {
		log.Error("Transaction error: " + err.Error())
	}

	stmt, err = tx.Prepare(insert)

	_, err = stmt.Exec(section.Name, section.Grace, section.Dataset, metadata.Name,
		metadata.Os, metadata.User, metadata.Group, metadata.Type, metadata.Link,
		metadata.Mtime, metadata.Ctime, metadata.Hash, metadata.Mode, section.Compressed)

	if err != nil {
		log.Error("Exec error: " + err.Error())
	}

	stmt.Close()
	tx.Commit()
}
