/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package server

import (
	"bitbucket.org/ebianchi/memento-common/common"
	"github.com/go-ini/ini"
	"github.com/op/go-logging"
	"os"
	"path/filepath"
	"server/database"
	"server/syncing"
	"strconv"
	"sync"
	"database/sql"
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
	var tx *sql.Tx
	var c = make(chan bool, POOL)
	var wg = new(sync.WaitGroup)

	var dataset, maxdatasets int
	var sections []*ini.Section

	sections = cfg.Sections()

	maxdatasets, _ = cfg.Section("dataset").Key(grace).Int()

	db.Open(log, cfg)
	defer db.Close()

	tx, _ = db.Conn.Begin()
	dataset = database.Getdataset(log, tx, grace)
	tx.Commit();

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
				sect := common.Section{
					Name:       section.Name(),
					Grace:      grace,
					Dataset:    dataset,
					Compressed: section.Key("compress").MustBool(),
				}

				go filesync(log, &sect, cfg, c, wg)
				c <- true
			}
		}
	}
	wg.Wait() // Wait for all the children to die
	close(c)

	tx, _ = db.Conn.Begin()
	database.Setdataset(log, tx, dataset, grace)
	tx.Commit()
}

func filesync(log *logging.Logger, section *common.Section, cfg *ini.File, c chan bool, wg *sync.WaitGroup) {
	defer func() {
		<-c
		wg.Done()
	}()

	Removedataset(log, cfg, section.Name, section.Grace, section.Dataset)

	log.Info("About to execute section " + section.Name)
	syncing.Filesync(log, section, cfg)
}

func Removedataset(log *logging.Logger, cfg *ini.File, section, grace string, dataset int) {
	var db database.DB

	log.Debug("About to delete dataset " + strconv.Itoa(dataset) + " for section " + section + " and grace " + grace)

	db.Open(log, cfg)
	defer db.Close()

	err := database.Deldataset(log, &db, section, grace, dataset)
	if err != nil {
		log.Error(err.Error())
	}

	repository := cfg.Section("general").Key("repository").String() +
		string(filepath.Separator) +
		grace +
		string(filepath.Separator) +
		strconv.Itoa(dataset)

	if section != "" {
		repository = repository + string(filepath.Separator) + section
	}

	os.RemoveAll(repository)

	log.Debug("Dataset deleted")
}
