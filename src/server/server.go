/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package server

import (
	"fmt"
	"github.com/go-ini/ini"
	"server/database"
)

var SECT_RESERVED = []string{"DEFAULT", "general", "database", "dataset"}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Sync(cfg *ini.File, grace string) {
	var db database.DB

	var dataset, maxdatasets int
	var sections []*ini.Section

	db.Open(cfg)
	defer db.Close()

	sections = cfg.Sections()

	maxdatasets, _ = cfg.Section("dataset").Key(grace).Int()
	dataset = database.Getdataset(&db, grace)

	if nextds := dataset + 1; nextds > maxdatasets {
		dataset = 1
	} else {
		dataset = dataset + 1
	}

	fmt.Println("dataset:", dataset)

	for _, section := range sections {
		if !contains(SECT_RESERVED, section.Name()) {
			fmt.Println(section.Name()) // TODO: add code for process a section
		}
	}
}
