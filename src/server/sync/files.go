/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package sync

import (
	"bitbucket.org/ebianchi/memento-common/common"
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/go-ini/ini"
	"net"
	"time"
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

func fs_metadata() {
	// TODO: write code for getting file's metadata from client
}

func Filesync(section *Section) {
	var buff *bufio.Reader
	var cmd common.JSONMessage
	var conn net.Conn
	var err error
	var result []byte

	// Execute pre_command
	if section.Section.Key("pre_command").String() != "" {
		conn, err = getsocket(section.Section)
		if err != nil {
			//TODO: manage error
			fmt.Println("Error: " + err.Error())
			return
		}
		buff = bufio.NewReader(conn)

		cmd = common.JSONMessage{}
		cmd.Context = "system"
		cmd.Command.Name = "exec"
		cmd.Command.Value = section.Section.Key("pre_command").String()

		if err = cmd.Send(conn); err != nil {
			//TODO: manage error
			fmt.Println("Sending failed: " + err.Error())
		} else {
			result, err = buff.ReadBytes('\n')
			if err != nil {
				//TODO: manage error
				fmt.Println("Receive failed: " + err.Error())
			}
			fmt.Println(string(result))
		}
		conn.Close()
	}

	conn, err = getsocket(section.Section)
	fs_metadata()
	conn.Close()

	// Execute post_command
	if section.Section.Key("post_command").String() != "" {
		conn, err = getsocket(section.Section)
		if err != nil {
			//TODO: manage error
			fmt.Println("Error: " + err.Error())
			return
		}
		buff = bufio.NewReader(conn)

		cmd = common.JSONMessage{}
		cmd.Context = "system"
		cmd.Command.Name = "exec"
		cmd.Command.Value = section.Section.Key("post_command").String()

		if err = cmd.Send(conn); err != nil {
			//TODO: manage error
			fmt.Println("Sending failed: " + err.Error())
		} else {
			result, err = buff.ReadBytes('\n')
			if err != nil {
				//TODO: manage error
				fmt.Println("Receive failed: " + err.Error())
			}
			fmt.Println(string(result))
		}
		conn.Close()
	}
}
