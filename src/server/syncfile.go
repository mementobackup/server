/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package server

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"github.com/go-ini/ini"
	"net"
	"time"
)

func getsocket(cfg *ini.Section) net.Conn {
	var conn net.Conn
	var tcpAddr *net.TCPAddr
	var err error

	host := cfg.Key("host").String()
	port := cfg.Key("port").String()

	address := host + ":" + port
	tcpAddr, err = net.ResolveTCPAddr("tcp", address)

	if err != nil {
		// TODO: manage error
		fmt.Println(err.Error())
	}

	if cfg.Key("ssl").MustBool() {
		cert, err := tls.LoadX509KeyPair(cfg.Key("sslkey").String(), cfg.Key("sslprivate").String())
		if err != nil {
			//TODO: manage error
		}

		config := tls.Config{Certificates: []tls.Certificate{cert}}

		now := time.Now()
		config.Time = func() time.Time { return now }
		config.Rand = rand.Reader

		conn, err = tls.Dial("tcp", tcpAddr.String(), &config)

	} else {

		conn, err = net.Dial("tcp", tcpAddr.String())
		if err != nil {
			// TODO: manage error
			fmt.Println(err.Error())
		}
		fmt.Printf("%T\n", conn)
	}
	return conn
}

func fs_metadata() {
	// TODO: write code for getting file's metadata from client
}

func syncfile(section *Section) {
	var result []byte

	conn := getsocket(section.Section)
	defer conn.Close()

	buff := bufio.NewReader(conn)

	// Execute pre_command
	if section.Section.Key("pre_command").String() != "" {
		_, err := conn.Write([]byte(section.Section.Key("pre_command").String() + "\n"))
		result, err = buff.ReadBytes('\n')
		if err != nil {
			//TODO: manage error
			fmt.Println(err.Error())
		}
		fmt.Println(string(result))
	}

	fs_metadata()

	// Execute post_command
	if section.Section.Key("post_command").String() != "" {
		_, err := conn.Write([]byte(section.Section.Key("post_command").String()))
		result, err = buff.ReadBytes('\n')
		if err != nil {
			//TODO: manage error
		}
		fmt.Println(string(result))
	}
}
