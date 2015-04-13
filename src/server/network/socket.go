/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package network

import (
	"crypto/rand"
	"crypto/tls"
	"errors"
	"github.com/go-ini/ini"
	"net"
	"time"
)

func Getsocket(cfg *ini.Section) (net.Conn, error) {
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
