/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package restore

import (
	"github.com/go-ini/ini"
	"github.com/mementobackup/common/src/common"
	"github.com/op/go-logging"
	"net"
	"path/filepath"
	"server/database"
	"server/dataset"
	"server/network"
)

func Filerestore(log *logging.Logger, section *common.Section, cfg *ini.File) {
	var cmd common.JSONMessage
	var res common.JSONFile
	var db database.DB
	var err error

	db.Open(log, cfg)
	defer db.Close()

	cmd.Context = "file"
	cmd.Command.Name = "put"

	if cfg.Section(section.Name).Key("path").String() == "" {
		log.Debug("About to full restore")
		cmd.Command.ACL = cfg.Section(section.Name).Key("acl").MustBool()
		for _, item := range []string{"directory", "file", "symlink"} {
			for res = range database.Listitems(log, &db, section, item) {
				if cmd.Command.ACL {
					log.Debug("About to restore ACLs for " + res.Name)
					res.Acl, err = database.Getacls(log, &db, res.Name, section)
					if err != nil {
						log.Error("ACLs extraction error for " + res.Name)
						log.Debug("error: " + err.Error())
					}
				}

				cmd.Command.Element = res
				put(log, section, cfg, &cmd)
			}
		}
	} else {
		log.Debug("About to restore path " + cfg.Section(section.Name).Key("path").String())

		cmd.Command.Element, err = database.Getitem(log, &db, cfg.Section(section.Name).Key("path").String(), section)
		if err != nil {
			log.Error("Error when putting data to section " + section.Name)
			log.Debug("error: " + err.Error())
		} else {
			cmd.Command.ACL = cfg.Section(section.Name).Key("acl").MustBool()
			put(log, section, cfg, &cmd)
		}
	}
}

func put(log *logging.Logger, section *common.Section, cfg *ini.File, cmd *common.JSONMessage) {
	var conn net.Conn
	var err error

	conn, err = network.Getsocket(cfg.Section(section.Name))
	if err != nil {
		log.Error("Error when opening connection with section " + section.Name)
		log.Debug("error: " + err.Error())
		return
	}
	defer conn.Close()

	cmd.Send(conn)

	if cmd.Command.Element.Type == "file" {
		transfered := dataset.Path(cfg, section, false) + string(filepath.Separator) + cmd.Command.Element.Name
		if err := common.Sendfile(transfered, conn); err != nil {
			log.Debug("Error when sending file: ", err.Error())
		}

	}
}
