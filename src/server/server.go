/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package server

import (
	"github.com/go-ini/ini"
	"server/database"
    "server/sync"
)

var SECT_RESERVED = []string{"DEFAULT", "general", "database", "dataset"}

type Section struct {
	Section *ini.Section
	Grace   string
	Dataset int
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Sync(cfg *ini.File, grace string) {
	const POOL = 5
	var db database.DB
	var c = make(chan bool, POOL)
	var wg = new(sync.WaitGroup)

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

	wg.Add(len(sections) - len(SECT_RESERVED))
	for _, section := range sections {
		if !contains(SECT_RESERVED, section.Name()) {
			if section.Key("type").String() == "file" { // FIXME: useless?
				s := Section{
					section,
					grace,
					dataset,
				}

				go call(&s, c, wg)
				c <- true
			}
		}
	}
	wg.Wait() // Wait for all the children to die
	close(c)
}

func call(section *Section, c chan bool, wg *sync.WaitGroup) {
	defer func() {
		<-c
		wg.Done()
	}()
	sync.Filesync(section)
}
