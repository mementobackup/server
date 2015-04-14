/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package server

import (
	"github.com/go-ini/ini"
	"github.com/op/go-logging"
	"server/database"
	"server/syncing"
	"strconv"
	"sync"
	"bitbucket.org/ebianchi/memento-common/common"
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

func Sync(log *logging.Logger, cfg *ini.File, grace string) {
	const POOL = 5
	var db database.DB
	var c = make(chan bool, POOL)
	var wg = new(sync.WaitGroup)

	var dataset, maxdatasets int
	var sections []*ini.Section

	db.Open(log, cfg)
	defer db.Close()

	sections = cfg.Sections()

	maxdatasets, _ = cfg.Section("dataset").Key(grace).Int()
	dataset = database.Getdataset(&db, grace)

	if nextds := dataset + 1; nextds > maxdatasets {
		dataset = 1
	} else {
		dataset = dataset + 1
	}
	log.Info("Dataset processed: " + strconv.Itoa(dataset))

	wg.Add(len(sections) - len(SECT_RESERVED))
	for _, section := range sections {
		if !contains(SECT_RESERVED, section.Name()) {
			if section.Key("type").String() == "file" { // FIXME: useless?
				s := common.Section{
					section,
					grace,
					dataset,
				}

				go filesync(log, &s, c, wg)
				c <- true
			}
		}
	}
	wg.Wait() // Wait for all the children to die
	close(c)
}

func filesync(log *logging.Logger, section *common.Section, c chan bool, wg *sync.WaitGroup) {
	defer func() {
		<-c
		wg.Done()
	}()
	log.Info("About to execute section " + section.Section.Name())
	syncing.Filesync(log, section)
}
