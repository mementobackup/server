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
	"fmt"
	"github.com/op/go-logging"
	"net"
	"server/generic"
)

func exec_command(log *logging.Logger, section *generic.Section, command string) {
	var buff *bufio.Reader
	var conn net.Conn
	var cmd common.JSONMessage
	var err error
	var result []byte

	if section.Section.Key(command).String() != "" {
		conn, err = getsocket(section.Section)
		if err != nil {
			log.Error("Error when executing " + command + ": " + err.Error())
			return
		}
		buff = bufio.NewReader(conn)

		cmd = common.JSONMessage{}
		cmd.Context = "system"
		cmd.Command.Name = "exec"
		cmd.Command.Value = section.Section.Key(command).String()

		if err = cmd.Send(conn); err != nil {
			log.Error("Sending " + command + " failed: " + err.Error())
		} else {
			result, err = buff.ReadBytes('\n')
			if err != nil {
				log.Error("Receive " + command + "result failed: " + err.Error())
			}
			fmt.Println(string(result))
		}
		conn.Close()
		log.Debug("Executed " + command)
	}

}
