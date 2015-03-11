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
	"log"
)

func (db *DB) check() error {
	var counted int

	query := "SELECT COUNT(rdb$relation_name) " +
		"FROM rdb$relations WHERE " +
		"rdb$relation_name NOT LIKE 'RDB$%' " +
		"AND rdb$relation_name NOT LIKE 'MON$%'"

	err := db.Conn.QueryRow(query).Scan(&counted)
	if err != nil {
		log.Fatal("Error in check database structure: " + err.Error())
	}

	if counted > 0 {
		return nil
	} else {
		return errors.New(string(counted))
	}
}

func (db *DB) create() {
	var tx *sql.Tx
	var domains = []string{"CREATE DOMAIN BOOLEAN AS SMALLINT CHECK (value is null or value in (0, 1))"}
	var indexes = []string{
		"CREATE INDEX idx_attrs_1 ON attrs(area, grace, dataset)",
		"CREATE INDEX idx_attrs_2 ON attrs(area, grace, dataset, hash)",
		"CREATE INDEX idx_attrs_3 ON attrs(area, grace, dataset, type)",
		"CREATE INDEX idx_acls_1 ON acls(area, grace, dataset)",
	}

	var tables = []string{
		"CREATE TABLE status ( " +
			"grace VARCHAR(5) CHARACTER SET UTF8, " +
			" actual INTEGER, " +
			" last_run TIMESTAMP)",
		"CREATE TABLE attrs ( " +
			"area VARCHAR(30) CHARACTER SET UTF8, " +
			"grace VARCHAR(5) CHARACTER SET UTF8, " +
			"dataset INTEGER, " +
			"element VARCHAR(1024) CHARACTER SET UTF8, " +
			"os VARCHAR(32) CHARACTER SET UTF8, " +
			"username VARCHAR(50) CHARACTER SET UTF8, " +
			"groupname VARCHAR(50) CHARACTER SET UTF8, " +
			"type VARCHAR(9) CHARACTER SET UTF8, " +
			"link VARCHAR(1024) CHARACTER SET UTF8, " +
			"hash VARCHAR(32) CHARACTER SET UTF8, " +
			"perms VARCHAR(32) CHARACTER SET UTF8, " +
			"mtime BIGINT, " +
			"ctime BIGINT, " +
			"compressed BOOLEAN)",
		"CREATE TABLE acls ( " +
			"area VARCHAR(30) CHARACTER SET UTF8, " +
			"grace VARCHAR(5) CHARACTER SET UTF8, " +
			"dataset INTEGER, " +
			"element VARCHAR(1024) CHARACTER SET UTF8, " +
			"name VARCHAR(50) CHARACTER SET UTF8, " +
			"type VARCHAR(5) CHARACTER SET UTF8, " +
			"perms VARCHAR(3) CHARACTER SET UTF8)",
	}

	var data = []string{
		"INSERT INTO status VALUES('hour', 0, CURRENT_TIMESTAMP)",
		"INSERT INTO status VALUES('day', 0, CURRENT_TIMESTAMP)",
		"INSERT INTO status VALUES('week', 0, CURRENT_TIMESTAMP)",
		"INSERT INTO status VALUES('month', 0, CURRENT_TIMESTAMP)",
	}

	tx, _ = db.Conn.Begin()
	for _, query := range domains {
		_, err := tx.Exec(query)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
			break
		}
	}
	tx.Commit()

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
