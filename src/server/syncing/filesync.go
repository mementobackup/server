/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package syncing

import (
	"bitbucket.org/ebianchi/memento-common/common"
	"bufio"
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

	buff = bufio.NewReader(conn)
	for {
		res = common.JSONResult{}

		data, err = buff.ReadBytes('\n')

		if err != nil {
			if err == io.EOF {
				log.Debug("All data in connection are readed, exit")
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

		db.Open(log, cfg)
		defer db.Close()

		database.Saveattrs(log, &db, section, res.Data)
	}
}

func fs_get_data(log *logging.Logger, section *common.Section, cfg *ini.File) {
	var res common.JSONFile
	var db database.DB

	/*
		if section.Dataset - 1 == 0 {
			previous := cfg.Section("dataset").Key(section.Grace)
		} else {
			previous := section.Dataset -1
		}
	*/

	db.Open(log, cfg)
	defer db.Close()

	for _, item := range []string{"directory", "file", "symlink"} {
		for res = range database.Listitems(log, &db, section, item) {
			switch item {
				case "directory":
					fs_save_data(log, cfg, section, res, false)
				case "file":
				case "symlink":
			}
		}
	}
}

func Filesync(log *logging.Logger, section *common.Section, cfg *ini.File) {
	// Execute pre_command
	exec_command(log, cfg.Section(section.Name), "pre_command")

	// Retrieve file's metadata
	fs_get_metadata(log, section, cfg)

	// Retrieve file's metadata
	fs_get_data(log, section, cfg)

	// Execute post_command
	exec_command(log, cfg.Section(section.Name), "post_command")
}