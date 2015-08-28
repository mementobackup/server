/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package server

import (
	"bitbucket.org/ebianchi/memento-common/common"
	"bufio"
	"encoding/json"
	"github.com/go-ini/ini"
	"github.com/op/go-logging"
	"net"
	"server/network"
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

func exec_command(log *logging.Logger, section *ini.Section, command string) {
	var buff *bufio.Reader
	var conn net.Conn
	var cmd common.JSONMessage
	var res common.JSONResult
	var err error
	var result []byte

	if section.Key(command).String() != "" {
		conn, err = network.Getsocket(section)
		if err != nil {
			log.Error("Error when executing " + command + ": " + err.Error())
			return
		}
		defer conn.Close()

		buff = bufio.NewReader(conn)

		cmd = common.JSONMessage{}
		cmd.Context = "system"
		cmd.Command.Name = "exec"
		cmd.Command.Cmd = section.Key(command).String()

		if err = cmd.Send(conn); err != nil {
			log.Error("Sending " + command + " failed: " + err.Error())
		} else {
			result, err = buff.ReadBytes('\n')
			if err != nil {
				log.Error("Receive " + command + " result failed: " + err.Error())
			}

			err = json.Unmarshal(result, &res)
			if err != nil {
				log.Error("Responde result for " + command + " failed: " + err.Error())
			}

			if res.Result == "ok" {
				// TODO: manage corrected execution for remote command
			} else {
				log.Error("Error in " + command + " execution: " + res.Message)
			}
		}
		log.Debug("Executed " + command)
	}
}
