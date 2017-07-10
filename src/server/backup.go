/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package server

import (
	"database/sql"
	"github.com/go-ini/ini"
	"github.com/mementobackup/common/src/common"
	"github.com/op/go-logging"
	"server/backup"
	"server/database"
	"server/dataset"
	"strconv"
	"sync"
)

func Backup(log *logging.Logger, cfg *ini.File, grace string, reload bool) {
	const POOL = 5
	var sysdb database.DB
	var tx *sql.Tx
	var c = make(chan bool, POOL)
	var wg = new(sync.WaitGroup)

	var dataset, maxdatasets int
	var sections []*ini.Section

	sections = cfg.Sections()

	maxdatasets, _ = cfg.Section("dataset").Key(grace).Int()

	sysdb.Setlocation(cfg.Section("general").Key("repository").String())
	sysdb.Open(log)
	defer sysdb.Close()

	tx, _ = sysdb.Conn.Begin()
	dataset = database.GetDataset(log, tx, grace)
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

				go fileBackup(log, &sect, cfg, c, wg)
				c <- true
			}
		}
	}
	wg.Wait() // Wait for all the children to die
	close(c)

	tx, _ = sysdb.Conn.Begin()
	database.SetDataset(log, tx, dataset, grace)
	tx.Commit()
}

func fileBackup(log *logging.Logger, section *common.Section, cfg *ini.File, c chan bool, wg *sync.WaitGroup) {
	defer func() {
		<-c
		wg.Done()
	}()

	dataset.DelDataset(log, cfg, section.Name, section.Grace, section.Dataset)

	log.Info("About to backup section " + section.Name)

	// Execute pre_command
	if err := execCommand(log, cfg.Section(section.Name), "pre_command"); err != nil {
		log.Debug(err.Error())
		// TODO: manage error
	}

	// Backup!
	backup.Filebackup(log, section, cfg)

	// Execute post_command
	if err := execCommand(log, cfg.Section(section.Name), "post_command"); err != nil {
		log.Debug(err.Error())
		// TODO: manage error
	}
}
