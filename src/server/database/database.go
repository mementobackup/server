/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package database

import (
	"code.google.com/p/goconf/conf"
	"database/sql"
	_ "github.com/nakagami/firebirdsql"
	"log"
)

type DB struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
	Conn     *sql.DB
}

func (db *DB) populate(cfg *conf.ConfigFile) {
	db.User, _ = cfg.GetString("database", "user")
	db.Password, _ = cfg.GetString("database", "password")
	db.Host, _ = cfg.GetString("database", "host")
	db.Port, _ = cfg.GetString("database", "port")
	db.Database, _ = cfg.GetString("database", "dbname")
}

func (db *DB) Open(cfg *conf.ConfigFile) {
	var err error

	db.populate(cfg)
	db.Conn, err = sql.Open("firebirdsql", db.User+":"+db.Password+"@"+db.Host+":"+db.Port+"/"+db.Database)

	if err != nil {
		log.Fatal(err)
	}

	err = db.Conn.Ping()
	if err != nil {
		log.Fatal(err)
	}

	if err = db.check(); err != nil {
		db.create()
	}
}

func (db *DB) Close() {
	db.Conn.Close()
}
