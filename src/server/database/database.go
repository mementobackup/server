/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package database

import (
	"database/sql"
	_ "github.com/gwenn/gosqlite"
	"github.com/op/go-logging"
)

type DB struct {
	Location string
	Conn     *sql.DB
}

func (db *DB) Open(log *logging.Logger, location string) {
	var err error
	db.Location = location

	db.Conn, err = sql.Open("sqlite3", db.Location)

	if err != nil {
		log.Fatal(err)
	}

	err = db.Conn.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Connection with database opened")

	if err = db.check(log); err != nil {
		db.create(log)
		log.Debug("Created database schema")
	}
}

func (db *DB) Close() {
	db.Conn.Close()
}
