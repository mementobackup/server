/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package syncing

import (
	"bitbucket.org/ebianchi/memento-common/common"
	"github.com/op/go-logging"
	"net"
	"server/network"
	"strings"
)

func fs_metadata(log *logging.Logger, section *common.Section) {
	var conn net.Conn
	var cmd common.JSONMessage
	var err error

	log.Debug("Getting metadata for " + section.Grace)

	cmd = common.JSONMessage{}
	cmd.Context = "file"
	cmd.Command.Name = "list"
	cmd.Command.Directory = strings.Split(section.Section.Key("path").String(), ",")
	cmd.Command.ACL = section.Section.Key("acl").MustBool()

	conn, err = network.Getsocket(section.Section)
	if err != nil {
		log.Error("Error when getting files metadata for section " + section.Section.Name())
		log.Debug("Files metadata error: " + err.Error())
		return
	}
	defer conn.Close()

	// TODO: write code for getting file's metadata from client
}

func Filesync(log *logging.Logger, section *common.Section) {
	// Execute pre_command
	exec_command(log, section, "pre_command")

	fs_metadata(log, section)

	// Execute post_command
	exec_command(log, section, "post_command")
}
