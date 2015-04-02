/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package database

import (
	"database/sql"
	"github.com/go-ini/ini"
	_ "github.com/nakagami/firebirdsql"
	"github.com/op/go-logging"
)

type DB struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
	Conn     *sql.DB
}

func (db *DB) populate(cfg *ini.File) {
	db.User = cfg.Section("database").Key("user").String()
	db.Password = cfg.Section("database").Key("password").String()
	db.Host = cfg.Section("database").Key("host").String()
	db.Port = cfg.Section("database").Key("port").String()
	db.Database = cfg.Section("database").Key("dbname").String()
}

func (db *DB) Open(log *logging.Logger, cfg *ini.File) {
	var err error

	db.populate(cfg)
	db.Conn, err = sql.Open("firebirdsql", db.User+":"+db.Password+"@"+db.Host+":"+db.Port+"/"+db.Database)

	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Connection with database instantied")

	err = db.Conn.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Connection with database opened")

	if err = db.check(); err != nil {
		db.create()
		log.Debug("Created database schema")
	}
}

func (db *DB) Close() {
	db.Conn.Close()
}
