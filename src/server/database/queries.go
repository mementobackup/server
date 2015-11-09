/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package database

import (
	"database/sql"
	"errors"
	"github.com/mementobackup/common/src/common"
	"github.com/op/go-logging"
	"strconv"
)

func errorMsg(log *logging.Logger, position int, message string) *common.OperationErr {
	log.Error(message)
	return &common.OperationErr{Operation: "database",
		Position: position,
		Message:  message,
	}
}

func Saveacls(log *logging.Logger, tx *sql.Tx, section *common.Section, element string, acls []common.JSONFileAcl) error {
	var acl common.JSONFileAcl
	var stmt *sql.Stmt
	var err error

	var insert = "INSERT INTO acls" +
		"(area, grace, dataset, element, name, type, perms)" +
		" VALUES($1, $2, $3, $4, $5, $6, $7)"

	stmt, err = tx.Prepare(insert)
	if err != nil {
		log.Debug("Failed prepare: " + err.Error())
		return errorMsg(log, 1, "Cannot save ACLs for element "+section.Name)
	}

	for _, acl = range acls {
		if acl.User != "" {
			_, err = stmt.Exec(section.Name, section.Grace, section.Dataset, element,
				acl.User, "user", acl.Mode)
		} else {
			_, err = stmt.Exec(section.Name, section.Grace, section.Dataset, element,
				acl.Group, "group", acl.Mode)
		}
		if err != nil {
			log.Debug("Failed execute: " + err.Error())
			return errorMsg(log, 2, "Cannot save ACLs for element "+section.Name)
		}
	}

	stmt.Close()
	return nil
}

func Saveattrs(log *logging.Logger, tx *sql.Tx, section *common.Section, metadata common.JSONFile) error {
	var stmt *sql.Stmt
	var err error

	var insert = "INSERT INTO attrs" +
		"(area, grace, dataset, element, os, username, groupname, type," +
		" link, mtime, ctime, hash, perms, compressed)" +
		" VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)"

	stmt, err = tx.Prepare(insert)
	if err != nil {
		log.Debug("Failed prepare: " + err.Error())
		return errorMsg(log, 3, "Cannot save Attributes for element "+section.Name)
	}

	_, err = stmt.Exec(section.Name, section.Grace, section.Dataset, metadata.Name,
		metadata.Os, metadata.User, metadata.Group, metadata.Type, metadata.Link,
		metadata.Mtime, metadata.Ctime, metadata.Hash, metadata.Mode, section.Compressed)

	if err != nil {
		log.Debug("Failed execute: " + err.Error())
		return errorMsg(log, 4, "Cannot save Attributes for element "+section.Name)
	}

	stmt.Close()
	return nil
}

func Listitems(log *logging.Logger, db *DB, section *common.Section, itemtype string) <-chan common.JSONFile {
	var resitem common.JSONFile
	var rows *sql.Rows
	var element, os, hash, link string
	var err error

	var query = "SELECT element, os, hash, link" +
		" FROM attrs WHERE type = $1 AND area = $2 AND grace = $3 AND dataset = $4"

	result := make(chan common.JSONFile)

	rows, err = db.Conn.Query(query, itemtype, section.Name, section.Grace, section.Dataset)
	if err != nil {
		log.Error("List items error: " + err.Error())
	}

	// Return a generator
	go func() {
		for rows.Next() {
			err = rows.Scan(&element, &os, &hash, &link)
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

func Itemexist(log *logging.Logger, db *DB, item *common.JSONFile, section *common.Section, previous int) bool {
	var dataset, result int

	if previous > 0 {
		dataset = previous
	} else {
		dataset = section.Dataset
	}

	var query = "SELECT count(element) FROM attrs" +
		" WHERE element = $1 AND hash = $2" +
		" AND area = $3 AND grace = $4 AND dataset = $5"

	log.Debug("Searching if item " + item.Name + " exists in dataset " + strconv.Itoa(dataset))
	err := db.Conn.QueryRow(query,
		item.Name, item.Hash, section.Name, section.Grace, dataset).Scan(&result)
	if err != nil {
		log.Debug("Error when finding item: " + err.Error())
		return false
	}

	if result > 0 {
		return true
	} else {
		return false
	}
}

// Getitem function return item's attributes in element struct, otherwise return an error
func Getitem(log *logging.Logger, db *DB, element *common.JSONFile, section *common.Section) error {
	// TODO: extract ACLs for item
	var rows *sql.Rows
	var err error

	var query = "SELECT os, username, groupname, type, " +
		"link, mtime, ctime, hash, perms, compressed FROM attrs " +
		"WHERE element = $1 AND area = $2 AND grace = $3 AND dataset = $4"

	rows, err = db.Conn.Query(query, element.Name, section.Name, section.Grace, section.Dataset)
	if err != nil {
		log.Error("Get item error: " + err.Error())
	}

	rows.Next()
	err = rows.Scan(&element.Os, &element.User, &element.Group, &element.Type, &element.Link,
		&element.Mtime, &element.Ctime, &element.Hash, &element.Mode, &element.Compressed)
	if err != nil {
		log.Error("Get item values extraction error: " + err.Error())
	} else {
		err = Getacls(log, db, element, section)
	}

	return err
}

// Getacls function return item's ACLs in element struct, otherwise return an error
func Getacls(log *logging.Logger, db *DB, element *common.JSONFile, section *common.Section) error {
	var rows *sql.Rows
	var data common.JSONFileAcl
	var aclname, acltype, aclperms string
	var err error

	var query = "SELECT name, type, perms FROM acls " +
	"WHERE element = $1 AND area = $2 AND grace = $3 AND dataset = $4"

	rows, err = db.Conn.Query(query, element.Name, section.Name, section.Grace, section.Dataset)
	if err != nil {
		log.Error("Get acls error: " + err.Error())
	} else {
		for rows.Next() {
			err = rows.Scan(&aclname, &acltype, &aclperms)
			if err != nil {
				log.Error("List values extraction error: " + err.Error())
			} else {
				if acltype == "user" {
					data.User = aclname
					data.Group = ""
				} else {
					data.Group = aclname
					data.User = ""
				}
				data.Mode = aclperms
				element.Acl = append(element.Acl, data)
			}
		}
	}

	return err
}

func Getdataset(log *logging.Logger, tx *sql.Tx, grace string) int {
	var result int

	var query = "SELECT actual FROM status WHERE grace = $1"

	err := tx.QueryRow(query, grace).Scan(&result)
	if err != nil {
		log.Debug("Error when getting dataset: " + err.Error())
		return 0
	}
	return result
}

func Setdataset(log *logging.Logger, tx *sql.Tx, actual int, grace string) error {
	var stmt *sql.Stmt
	var err error

	var query = "UPDATE status SET actual = $1, last_run = CURRENT_TIMESTAMP WHERE grace = $2"

	stmt, err = tx.Prepare(query)
	if err != nil {
		log.Debug("Failed prepare: " + err.Error())
		return errorMsg(log, 5, "Error when setting actual dataset")
	}

	_, err = stmt.Exec(actual, grace)
	if err != nil {
		log.Debug("Failed prepare: " + err.Error())
		return errorMsg(log, 6, "Error when setting actual dataset")
	}

	log.Debug("Dataset updated: " + strconv.Itoa(actual))

	stmt.Close()

	return nil
}

func Deldataset(log *logging.Logger, db *DB, section, grace string, dataset int) error {
	var tx *sql.Tx
	var stmt *sql.Stmt
	var err error

	var tables = []string{"attrs", "acls"}
	var queries = []string{
		"DELETE FROM " + tables[0] + " WHERE grace = $1 AND dataset = $2",
		"DELETE FROM " + tables[1] + " WHERE grace = $1 AND dataset = $2",
	}

	var geterror = func(debugmessage, message string) error {
		tx.Rollback()

		log.Debug(debugmessage)
		return errors.New(message)
	}

	if section != "" {
		queries[0] = queries[0] + " AND area = $3"
		queries[1] = queries[1] + " AND area = $3"
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
