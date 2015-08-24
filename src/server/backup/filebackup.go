/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package backup

import (
	"bitbucket.org/ebianchi/memento-common/common"
	"bufio"
	"database/sql"
	"encoding/json"
	"github.com/go-ini/ini"
	"github.com/op/go-logging"
	"io"
	"net"
	"server/database"
	"server/network"
	"strings"
)

func fs_get_metadata(log *logging.Logger, section *common.Section, cfg *ini.File) {
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
	cmd.Command.Directory = strings.Split(cfg.Section(section.Name).Key("path").String(), ",")
	cmd.Command.ACL = cfg.Section(section.Name).Key("acl").MustBool()

	conn, err = network.Getsocket(cfg.Section(section.Name))
	if err != nil {
		log.Error("Error when opening connection with section " + section.Name)
		log.Debug("error: " + err.Error())
		return
	}
	defer conn.Close()

	cmd.Send(conn)

	db.Open(log, cfg)
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

		if err = database.Saveattrs(log, tx, section, res.Data); err != nil {
			log.Error("Failed saving database item: " + res.Data.Name)
			log.Debug("Trace: " + err.Error())
			tx.Rollback()

			return
		}

		if err = database.Saveacls(log, tx, section, res.Data.Name, res.Data.Acl); err != nil {
			log.Error("Failed saving ACLs into database for item: " + res.Data.Name)
			log.Debug("Trace: " + err.Error())
			tx.Rollback()

			return
		}
	}

	tx.Commit()
}

func fs_get_data(log *logging.Logger, section *common.Section, cfg *ini.File) {
	var previous int
	var res common.JSONFile
	var db database.DB

	if section.Dataset-1 == 0 {
		previous = cfg.Section("dataset").Key(section.Grace).MustInt()
	} else {
		previous = section.Dataset - 1
	}

	db.Open(log, cfg)
	defer db.Close()

	for _, item := range []string{"directory", "file", "symlink"} {
		for res = range database.Listitems(log, &db, section, item) {
			switch item {
			case "directory":
				fs_save_data(log, cfg, section, res, false)
			case "symlink":
				fs_save_data(log, cfg, section, res, false)
			case "file":
				if database.Itemexist(log, &db, &res, section, previous) {
					fs_save_data(log, cfg, section, res, true)
				} else {
					fs_save_data(log, cfg, section, res, false)
				}
			}
		}
	}
}

func Filebackup(log *logging.Logger, section *common.Section, cfg *ini.File) {
	// Execute pre_command
	exec_command(log, cfg.Section(section.Name), "pre_command")

	// Retrieve file's metadata
	fs_get_metadata(log, section, cfg)

	// Retrieve file's metadata
	fs_get_data(log, section, cfg)

	// Execute post_command
	exec_command(log, cfg.Section(section.Name), "post_command")
}
