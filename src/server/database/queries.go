/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package database

import (
	"database/sql"
	"github.com/mementobackup/common/src/common"
	"github.com/op/go-logging"
	"strconv"
)

func errorMsg(log *logging.Logger, position int, message string) *common.OperationErr {
	log.Error(message)
	return &common.OperationErr{
		Operation: "database",
		Position:  position,
		Message:   message,
	}
}

func SaveAcls(log *logging.Logger, tx *sql.Tx, section *common.Section, element string, acls []common.JSONFileAcl) error {
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

func SaveAttrs(log *logging.Logger, tx *sql.Tx, section *common.Section, metadata common.JSONFile) error {
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

func ListItems(log *logging.Logger, db *DB, section *common.Section, itemtype string) <-chan common.JSONFile {
	var resitem common.JSONFile
	var rows *sql.Rows
	var err error

	var query = "SELECT element, os, hash, link," +
		" username, groupname, type," +
		" mtime, ctime, perms, compressed" +
		" FROM attrs WHERE type = $1 AND area = $2 AND grace = $3 AND dataset = $4" +
		" ORDER BY element"

	result := make(chan common.JSONFile)

	rows, err = db.Conn.Query(query, itemtype, section.Name, section.Grace, section.Dataset)
	if err != nil {
		log.Error("List items error: " + err.Error())
	}

	// Return a generator
	go func() {
		for rows.Next() {
			err = rows.Scan(&resitem.Name, &resitem.Os, &resitem.Hash, &resitem.Link,
				&resitem.User, &resitem.Group, &resitem.Type, &resitem.Mtime,
				&resitem.Ctime, &resitem.Mode, &resitem.Compressed)
			if err != nil {
				log.Error("List values extraction error: " + err.Error())
			}

			result <- resitem

		}
		rows.Close()
		close(result)
	}()

	return result
}

func ItemExist(log *logging.Logger, db *DB, item *common.JSONFile, section *common.Section) bool {
	var result int

	var query = "SELECT count(element) FROM attrs" +
		" WHERE element = $1 AND hash = $2" +
		" AND area = $3 AND grace = $4"

	log.Debug("Searching if item " + item.Name + " exists in database " + db.Location)
	err := db.Conn.QueryRow(query,
		item.Name, item.Hash, section.Name, section.Grace).Scan(&result)
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

func GetItem(log *logging.Logger, db *DB, element string, section *common.Section) (common.JSONFile, error) {
	var result common.JSONFile
	var rows *sql.Rows
	var err error

	var query = "SELECT os, username, groupname, type, " +
		"link, mtime, ctime, hash, perms, compressed FROM attrs " +
		"WHERE element = $1 AND area = $2 AND grace = $3 AND dataset = $4"

	rows, err = db.Conn.Query(query, element, section.Name, section.Grace, section.Dataset)
	if err != nil {
		log.Error("Get item error: " + err.Error())
	}

	// FIXME: ugly. Find a method for using itemexists() for check if item exists or not
	if rows.Next() {
		err = rows.Scan(&result.Os, &result.User, &result.Group, &result.Type, &result.Link,
			&result.Mtime, &result.Ctime, &result.Hash, &result.Mode, &result.Compressed)
		if err != nil {
			log.Error("Get item values extraction error: " + err.Error())
		} else {
			result.Acl, err = GetAcls(log, db, element, section)
			if err != nil {
				log.Error("Get ACLs values extraction error: " + err.Error())
			} else {
				result.Name = element
			}
		}

		return result, err
	} else {
		return common.JSONFile{}, errorMsg(log, 7, "Item not found: "+element)
	}
}

func GetAcls(log *logging.Logger, db *DB, element string, section *common.Section) ([]common.JSONFileAcl, error) {
	var rows *sql.Rows
	var result []common.JSONFileAcl
	var data common.JSONFileAcl
	var aclname, acltype, aclperms string
	var err error

	var query = "SELECT name, type, perms FROM acls " +
		"WHERE element = $1 AND area = $2 AND grace = $3 AND dataset = $4"

	rows, err = db.Conn.Query(query, element, section.Name, section.Grace, section.Dataset)
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
				result = append(result, data)
			}
		}
	}

	return result, err
}

func GetDataset(log *logging.Logger, tx *sql.Tx, grace string) int {
	var result int

	var query = "SELECT actual FROM status WHERE grace = $1"

	err := tx.QueryRow(query, grace).Scan(&result)
	if err != nil {
		log.Debug("Error when getting dataset: " + err.Error())
		return 0
	}
	return result
}

func SetDataset(log *logging.Logger, tx *sql.Tx, actual int, grace string) error {
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