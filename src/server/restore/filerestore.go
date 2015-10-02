/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package restore

import (
	"bitbucket.org/ebianchi/memento-common/common"
	"fmt"
	"github.com/go-ini/ini"
	"github.com/op/go-logging"
	"net"
	"server/database"
	"server/network"
)

func Filerestore(log *logging.Logger, section *common.Section, cfg *ini.File) {
	var cmd common.JSONMessage
	var res common.JSONFile
	var db database.DB
	var compressed bool

	db.Open(log, cfg)
	defer db.Close()

	cmd.Context = "file"
	cmd.Command.Name = "put"

	if cfg.Section(section.Name).Key("path").String() == "" {
		log.Debug("About to full restore")
		cmd.Command.ACL = cfg.Section(section.Name).Key("acl").MustBool()
		// TODO: write full section restore
	} else {
		log.Debug("About to restore path " + cfg.Section(section.Name).Key("path").String())

		res, compressed = database.Getitem(log, &db, cfg.Section(section.Name).Key("path").String(), section)
		fmt.Println(res)
		fmt.Println(compressed)

		cmd.Command.Element = res
		cmd.Command.ACL = cfg.Section(section.Name).Key("acl").MustBool()
		put(log, section, cfg, &cmd)
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
}
