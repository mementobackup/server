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
	var compressed int
	var err error

	var insert = "INSERT INTO attrs" +
		"(area, grace, dataset, element, os, username, groupname, type," +
		" link, mtime, ctime, hash, perms, compressed)" +
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

	// Because Firebird doesn't support BOOLEAN values, and the driver
	// doesn't convert automatically to int, I need to convert boolean
	// value to int value
	if section.Compressed {
		compressed = 1
	} else {
		compressed = 0
	}

	tx, err = db.Conn.Begin()
	if err != nil {
		log.Error("Transaction error: " + err.Error())
		return
	}

	stmt, err = tx.Prepare(insert)

	_, err = stmt.Exec(section.Name, section.Grace, section.Dataset, metadata.Name,
		metadata.Os, metadata.User, metadata.Group, metadata.Type, metadata.Link,
		metadata.Mtime, metadata.Ctime, metadata.Hash, metadata.Mode, compressed)

	if err != nil {
		log.Error("Exec error: " + err.Error())
		tx.Rollback()
	}

	stmt.Close()
	tx.Commit()
}
