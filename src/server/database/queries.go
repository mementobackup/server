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
	"errors"
	"github.com/op/go-logging"
)

func saveacls(log *logging.Logger, tx *sql.Tx, section *common.Section, element string, acls []common.JSONFileAcl) {
	var acl common.JSONFileAcl
	var stmt *sql.Stmt
	var err error

	var insert = "INSERT INTO acls" +
		"(area, grace, dataset, element, name, type, perms)" +
		"VALUES(?, ?, ?, ?, ?, ?, ?)"

	stmt, err = tx.Prepare(insert)
	if err != nil {
		log.Error("Cannot save ACLs for element " + section.Name)
		log.Debug("Failed prepare: " + err.Error())
		return
	}

	for _, acl = range acls {
		if acl.User != "" {
			_, err = stmt.Exec(section.Name, section.Grace, section.Dataset, element,
				acl.User, "user", acl.Mode)
		} else {
			_, err = stmt.Exec(section.Name, section.Grace, section.Dataset, element,
				acl.Group, "user", acl.Mode)
		}
		if err != nil {
			log.Error("Cannot save ACLs for element " + section.Name)
			log.Debug("Failed execute: " + err.Error())
			tx.Rollback()

			break
		}
	}

	stmt.Close()
	tx.Commit()
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

	tx, err = db.Conn.Begin()
	if err != nil {
		log.Error("Transaction error: " + err.Error())
		return
	}

	saveacls(log, tx, section, metadata.Name, metadata.Acl)
}

func Listitems(log *logging.Logger, db *DB, section *common.Section, item string) <-chan common.JSONFile {
	var resitem common.JSONFile
	var rows *sql.Rows
	var element, os, hash, itemtype, link string
	var err error

	var query = "SELECT element, os, hash, type, link" +
		" FROM attrs WHERE type = ? AND area = ? AND grace = ? AND dataset = ?"

	result := make(chan common.JSONFile)

	rows, err = db.Conn.Query(query, item, section.Name, section.Grace, section.Dataset)
	if err != nil {
		log.Error("List items error: " + err.Error())
	}

	// Return a generator
	go func() {
		for rows.Next() {
			err = rows.Scan(&element, &os, &hash, &itemtype, &link)
			if err != nil {
				log.Error("List values extraction error: " + err.Error())
			}

			resitem = common.JSONFile{
				Name: element,
				Os:   os,
				Hash: hash,
				Type: itemtype,
				Link: link,
			}

			result <- resitem

		}
		rows.Close()
		close(result)
	}()

	return result
}

func Getdataset(log *logging.Logger, db *DB, grace string) int {
	var result int

	var query = "SELECT actual FROM status WHERE grace = ?"

	err := db.Conn.QueryRow(query, grace).Scan(&result)
	if err != nil {
		log.Debug("Error when getting dataset: " + err.Error())
		return 0
	}
	return result
}

func Deldataset(log *logging.Logger, db *DB, section, grace string, dataset int) error {
	var tx *sql.Tx
	var stmt *sql.Stmt
	var err error

	var tables = []string{"attrs", "acls"}
	var queries = []string{
		"DELETE FROM " + tables[0] + " WHERE grace = ? AND dataset = ?",
		"DELETE FROM " + tables[1] + " WHERE grace = ? AND dataset = ?",
	}

	var geterror = func(debugmessage, message string) error {
		tx.Rollback()

		log.Debug(debugmessage)
		return errors.New(message)
	}

	if section != "" {
		queries[0] = queries[0] + " AND area = ?"
		queries[1] = queries[1] + " AND area = ?"
	}

	tx, err = db.Conn.Begin()
	if err != nil {
		return geterror("Transaction error: "+err.Error(), "Problems with opening transaction")
	}

	for item, query := range queries {
		log.Debug("Delete table " + tables[item])

		stmt, err = tx.Prepare(query)
		if err != nil {
			return geterror("Error in prepare: "+err.Error(), "Problems when preparing query")
		}

		if section != "" {
			_, err = stmt.Exec(grace, dataset, section)
		} else {
			_, err = stmt.Exec(grace, dataset)
		}

		if err != nil {
			return geterror("Exec error: "+err.Error(), "Delete of dataset "+tables[item]+" wasn't possible")
		}
		stmt.Close()
	}
	tx.Commit()

	return nil
}
