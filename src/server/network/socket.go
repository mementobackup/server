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
	"crypto/x509"
	"errors"
	"github.com/go-ini/ini"
	"io/ioutil"
	"net"
	"time"
)

func GetSocket(cfg *ini.Section) (net.Conn, error) {
	var conn net.Conn
	var err error

	host := cfg.Key("host").String()
	port := cfg.Key("port").String()

	address := host + ":" + port
	if _, err = net.ResolveTCPAddr("tcp", address); err != nil {
		return nil, err
	}

	if cfg.Key("ssl").MustBool() {
		key := cfg.Key("sslcert").String()
		private := cfg.Key("sslkey").String()

		cert, err := tls.LoadX509KeyPair(key, private)
		if err != nil {
			err = errors.New("Error when loading SSL certificate: " + err.Error())
			return nil, err
		}

		caCert, err := ioutil.ReadFile(key)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		config := tls.Config{
			RootCAs:      caCertPool,
			Certificates: []tls.Certificate{cert},
		}

		now := time.Now()
		config.Time = func() time.Time { return now }
		config.Rand = rand.Reader

		conn, err = tls.Dial("tcp", address, &config)
		if err != nil {
			return nil, err
		}

	} else {
		conn, err = net.Dial("tcp", address)
		if err != nil {
			return nil, err
		}
	}

	return conn, err
}
