/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package server

import (
	"bufio"
	"encoding/json"
	"github.com/go-ini/ini"
	"github.com/mementobackup/common/src/common"
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

func errorMsg(log *logging.Logger, position int, message string) *common.OperationErr {
	log.Error(message)
	return &common.OperationErr{Operation: "exec",
		Position: position,
		Message:  message,
	}
}

func execCommand(log *logging.Logger, section *ini.Section, command string) error {
	var buff *bufio.Reader
	var conn net.Conn
	var cmd common.JSONMessage
	var res common.JSONResult
	var err error
	var result []byte

	if section.Key(command).String() != "" {
		conn, err = network.GetSocket(section)
		if err != nil {
			return errorMsg(log, 1, "Connection with "+command+" failed: "+err.Error())
		}
		defer conn.Close()

		buff = bufio.NewReader(conn)

		cmd = common.JSONMessage{}
		cmd.Context = "system"
		cmd.Command.Name = "exec"
		cmd.Command.Cmd = section.Key(command).String()

		if err = cmd.Send(conn); err != nil {
			return errorMsg(log, 2, "Sending "+command+" failed: "+err.Error())
		} else {
			result, err = buff.ReadBytes('\n')
			if err != nil {
				log.Error("Receive " + command + " result failed: " + err.Error())
			}

			err = json.Unmarshal(result, &res)
			if err != nil {
				msg := "Response result for " + command + " failed: " + err.Error()
				log.Error(msg)

				return errorMsg(log, 3, msg)
			}

			if res.Result == "ok" {
				log.Debug("Executed " + command)
				return nil
			} else {
				return errorMsg(log, 4, "Error in "+command+" execution: "+res.Message)
			}
		}
	} else {
		return nil
	}
}
