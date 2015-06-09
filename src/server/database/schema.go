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
	"github.com/op/go-logging"
)

func (db *DB) check(log *logging.Logger) error {
	var counted int

	query := "SELECT Count(*) FROM information_schema.tables " +
	"WHERE table_schema = 'public' " +
	"AND table_name in ('status', 'attrs', 'acls')"

	err := db.Conn.QueryRow(query).Scan(&counted)
	if err != nil {
		log.Fatal("Schema 1: Error in check database structure: " + err.Error())
	}

	if counted > 0 {
		return nil
	} else {
		return errors.New(string(counted))
	}
}

func (db *DB) create(log *logging.Logger) {
	var tx *sql.Tx

	var indexes = []string{
		"CREATE INDEX idx_attrs_1 ON attrs(area, grace, dataset)",
		"CREATE INDEX idx_attrs_2 ON attrs(area, grace, dataset, hash)",
		"CREATE INDEX idx_attrs_3 ON attrs(area, grace, dataset, type)",
		"CREATE INDEX idx_acls_1 ON acls(area, grace, dataset)",
	}

	var tables = []string{
		"CREATE TABLE status ( " +
		"grace VARCHAR(5), " +
		" actual INTEGER, " +
		" last_run TIMESTAMP)",
		"CREATE TABLE attrs ( " +
		"area VARCHAR(30), " +
		"grace VARCHAR(5), " +
		"dataset INTEGER, " +
		"element VARCHAR(1024), " +
		"os VARCHAR(32), " +
		"username VARCHAR(50), " +
		"groupname VARCHAR(50), " +
		"type VARCHAR(9), " +
		"link VARCHAR(1024), " +
		"hash VARCHAR(32), " +
		"perms VARCHAR(32), " +
		"mtime BIGINT, " +
		"ctime BIGINT, " +
		"compressed BOOLEAN)",
		"CREATE TABLE acls ( " +
		"area VARCHAR(30), " +
		"grace VARCHAR(5), " +
		"dataset INTEGER, " +
		"element VARCHAR(1024), " +
		"name VARCHAR(50), " +
		"type VARCHAR(5), " +
		"perms VARCHAR(3))",
	}

	var data = []string{
		"INSERT INTO status VALUES('hour', 0, CURRENT_TIMESTAMP)",
		"INSERT INTO status VALUES('day', 0, CURRENT_TIMESTAMP)",
		"INSERT INTO status VALUES('week', 0, CURRENT_TIMESTAMP)",
		"INSERT INTO status VALUES('month', 0, CURRENT_TIMESTAMP)",
	}

	tx, _ = db.Conn.Begin()
	for _, query := range tables {
		_, err := tx.Exec(query)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
			break
		}
	}
	tx.Commit()

	tx, _ = db.Conn.Begin()
	for _, query := range indexes {
		_, err := tx.Exec(query)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
			break
		}
	}
	tx.Commit()

	tx, _ = db.Conn.Begin()
	for _, query := range data {
		_, err := tx.Exec(query)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
			break
		}
	}
	tx.Commit()
}
