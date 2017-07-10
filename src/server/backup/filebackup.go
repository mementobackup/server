/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package backup

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"github.com/go-ini/ini"
	"github.com/mementobackup/common/src/common"
	"github.com/op/go-logging"
	"io"
	"net"
	"server/database"
	"server/network"
	"strconv"
	"strings"
)

func fsGetMetadata(log *logging.Logger, section *common.Section, cfg *ini.File) {
	var conn net.Conn
	var cmd common.JSONMessage
	var db database.DB
	var buff *bufio.Reader
	var res common.JSONResult
	var tx *sql.Tx
	var data []byte
	var err error

	log.Debug("Getting metadata for " + section.Name)

	cmd.Context = "file"
	cmd.Command.Name = "list"
	cmd.Command.Paths = strings.Split(cfg.Section(section.Name).Key("path").String(), ",")
	cmd.Command.ACL = cfg.Section(section.Name).Key("acl").MustBool()
	cmd.Command.Exclude = cfg.Section(section.Name).Key("exclude").String()
	log.Debug("Metadata command request: %+v", cmd)

	conn, err = network.GetSocket(cfg.Section(section.Name))
	if err != nil {
		log.Error("Error when opening connection with section " + section.Name)
		log.Debug("error: " + err.Error())
		return
	}
	defer conn.Close()

	cmd.Send(conn)

	db.Setlocation("", cfg.Section("general").Key("repository").String(),
		section.Grace,
		strconv.Itoa(section.Dataset),
		section.Name)
	db.Open(log)
	defer db.Close()

	tx, err = db.Conn.Begin()
	if err != nil {
		log.Error("Transaction error for section " + section.Name)
		log.Debug("Trace: " + err.Error())

		return
	}

	buff = bufio.NewReader(conn)
	for {
		res = common.JSONResult{}

		data, err = buff.ReadBytes('\n')

		if err != nil {
			if err == io.EOF {
				log.Debug("All files's metadata are saved")
				break
			}

			log.Error("Error when getting files metadata for section " + section.Name)
			log.Debug("error: " + err.Error())
			return
		}

		err = json.Unmarshal(data, &res)
		if err != nil {
			log.Error("Error when parsing JSON data for section " + section.Name)
			log.Debug("error: " + err.Error())
			return
		}

		log.Debug("Metadata received: %+v", res)
		if err = database.SaveAttrs(log, tx, section, res.Data); err != nil {
			log.Error("Failed saving database item: " + res.Data.Name)
			log.Debug("Trace: " + err.Error())
			tx.Rollback()

			return
		}

		if err = database.SaveAcls(log, tx, section, res.Data.Name, res.Data.Acl); err != nil {
			log.Error("Failed saving ACLs into database for item: " + res.Data.Name)
			log.Debug("Trace: " + err.Error())
			tx.Rollback()

			return
		}
	}

	tx.Commit()
}

func fsGetData(log *logging.Logger, section *common.Section, cfg *ini.File) {
	var previous int
	var res common.JSONFile
	var curdb, olddb database.DB

	if section.Dataset-1 == 0 {
		previous = cfg.Section("dataset").Key(section.Grace).MustInt()
	} else {
		previous = section.Dataset - 1
	}

	curdb.Setlocation("", cfg.Section("general").Key("repository").String(),
		section.Grace,
		strconv.Itoa(section.Dataset),
		section.Name)
	olddb.Setlocation("", cfg.Section("general").Key("repository").String(),
		section.Grace,
		strconv.Itoa(previous),
		section.Name)

	curdb.Open(log)
	olddb.Open(log)

	defer curdb.Close()
	defer olddb.Close()

	for _, item := range []string{"directory", "file", "symlink"} {
		for res = range database.ListItems(log, &curdb, section, item) {
			switch item {
			case "directory":
				fsSaveData(log, cfg, section, res, false)
			case "symlink":
				fsSaveData(log, cfg, section, res, false)
			case "file":
				if database.ItemExist(log, &olddb, &res, section) {
					fsSaveData(log, cfg, section, res, true)
				} else {
					fsSaveData(log, cfg, section, res, false)
				}
			}
		}
	}
}

func Filebackup(log *logging.Logger, section *common.Section, cfg *ini.File) {
	// Retrieve file's metadata
	fsGetMetadata(log, section, cfg)

	// Retrieve file's data
	fsGetData(log, section, cfg)
}
