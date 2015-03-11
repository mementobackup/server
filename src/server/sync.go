/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package server

import (
	"code.google.com/p/goconf/conf"
	"fmt"
	"server/database"
)

var SECT_RESERVED = []string{"default", "general", "database", "dataset"}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Sync(cfg *conf.ConfigFile, grace string) {
	var db database.DB

	var dataset, maxdatasets int
	var sections []string

	db.Open(cfg)
	defer db.Close()

	sections = cfg.GetSections()

	maxdatasets, _ = cfg.GetInt("dataset", grace)
	dataset = database.Getdataset(&db, grace)

	if nextds := dataset + 1; nextds > maxdatasets {
		dataset = 1
	} else {
		dataset = dataset + 1
	}

	fmt.Println("dataset:", dataset)

	for _, section := range sections {
		if !contains(SECT_RESERVED, section) {
			fmt.Println(section) // TODO: add code for process a section
		}
	}
}
