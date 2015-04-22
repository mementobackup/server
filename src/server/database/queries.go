/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package database

import (
	"bitbucket.org/ebianchi/memento-common/common"
	"fmt"
	"database/sql"
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

func Saveattrs(db *DB, section *common.Section, metadata common.JSONFile) {
	var stmt *sql.Stmt
	var err error

	var insert = "INSERT INTO attrs" +
		"(area, grace, dataset, element, os, username, groupname, type," +
		" link, mtime, ctime, hash, perms, compressed)" +
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

	stmt, err = db.Conn.Prepare(insert)
	if err != nil {
		fmt.Println("Prepare error: " + err.Error())
	}
	defer stmt.Close()

	_, err = stmt.Exec(section.Name, section.Grace, section.Dataset, metadata.Name,
		metadata.Os, metadata.User, metadata.Group, metadata.Type, metadata.Link,
		metadata.Mtime, metadata.Ctime, metadata.Hash, metadata.Mode, section.Compressed)

	if err != nil {
		fmt.Println("Exec error: " + err.Error())
	}
}
