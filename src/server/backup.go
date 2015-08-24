/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package server

import (
	"bitbucket.org/ebianchi/memento-common/common"
	"database/sql"
	"github.com/go-ini/ini"
	"github.com/op/go-logging"
	"server/database"
	"server/dataset"
	"server/backup"
	"strconv"
	"sync"
)

func Backup(log *logging.Logger, cfg *ini.File, grace string, reload bool) {
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
	tx.Commit()

	if !reload {
		if nextds := dataset + 1; nextds > maxdatasets {
			dataset = 1
		} else {
			dataset = dataset + 1
		}
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

				go filebackup(log, &sect, cfg, c, wg)
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

func filebackup(log *logging.Logger, section *common.Section, cfg *ini.File, c chan bool, wg *sync.WaitGroup) {
	defer func() {
		<-c
		wg.Done()
	}()

	dataset.Deldataset(log, cfg, section.Name, section.Grace, section.Dataset)

	log.Info("About to execute section " + section.Name)
	backup.Filebackup(log, section, cfg)
}
