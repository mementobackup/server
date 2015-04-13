/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package syncing

import (
	"bitbucket.org/ebianchi/memento-common/common"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"github.com/go-ini/ini"
	"github.com/op/go-logging"
	"net"
	"server/generic"
	"time"
	"strings"
)

func getsocket(cfg *ini.Section) (net.Conn, error) {
	var conn net.Conn
	var tcpAddr *net.TCPAddr
	var err error

	host := cfg.Key("host").String()
	port := cfg.Key("port").String()

	address := host + ":" + port
	tcpAddr, err = net.ResolveTCPAddr("tcp", address)

	if err != nil {
		return nil, err
	}

	if cfg.Key("ssl").MustBool() {
		cert, err := tls.LoadX509KeyPair(cfg.Key("sslkey").String(), cfg.Key("sslprivate").String())
		if err != nil {
			err = errors.New("Error when loading SSL certificate: " + err.Error())
			return nil, err
		}

		config := tls.Config{Certificates: []tls.Certificate{cert}}

		now := time.Now()
		config.Time = func() time.Time { return now }
		config.Rand = rand.Reader

		conn, err = tls.Dial("tcp", tcpAddr.String(), &config)
		if err != nil {
			return nil, err
		}

	} else {
		conn, err = net.Dial("tcp", tcpAddr.String())
		if err != nil {
			return nil, err
		}
	}

	return conn, err
}

func fs_metadata(log *logging.Logger, section *generic.Section) {
	var conn net.Conn
	var cmd common.JSONMessage
	var err error

	log.Debug("Getting metadata for " + section.Grace)

	cmd = common.JSONMessage{}
	cmd.Context = "file"
	cmd.Command.Name = "list"
	cmd.Command.Directory = strings.Split(section.Section.Key("path").String(), ",")
	cmd.Command.ACL = section.Section.Key("acl").MustBool()

	conn, err = getsocket(section.Section)
	if err != nil {
		log.Error("Error when getting files metadata for section " + section.Section.Name())
		log.Debug("Files metadata error: " + err.Error())
		return
	}

	conn.Close()

	// TODO: write code for getting file's metadata from client
}

func Filesync(log *logging.Logger, section *generic.Section) {
	// Execute pre_command
	exec_command(log, section, "pre_command")

	fs_metadata(log, section)

	// Execute post_command
	exec_command(log, section, "post_command")
}
