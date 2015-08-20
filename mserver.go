/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package main

import (
	"fmt"
	"github.com/gaal/go-options/options"
	"github.com/go-ini/ini"
	"github.com/op/go-logging"
	"os"
	"path/filepath"
	"server"
	"strings"
)

const VERSION = "1.0"

const SPEC = `
Memento - A backup system
Usage: mserver [OPTIONS]
--
h,help                Print this help
v,version             Print version
b,backup=             Perform backup operation
r,restore=            Perform restore operation
H,hour                Hourly backup
D,day                 Daily backup
W,week                Weekly backup
M,month               Monthly backup
R,reload-dataset      Reload last dataset
`

func check_structure(repository string) {
	directories := []string{
		"hour",
		"day",
		"week",
		"month",
	}

	if _, err := os.Stat(repository); err != nil {
		for _, dir := range directories {
			if os.IsNotExist(err) {
				os.MkdirAll(repository+string(filepath.Separator)+dir, 0755)
			}
		}
	}
}

func setlog(level logging.Level, filename string) *logging.Logger {
	var backend *logging.LogBackend
	var log = logging.MustGetLogger("Memento Server")
	var format logging.Formatter

	if strings.ToUpper(level.String()) != "DEBUG" {
		format = logging.MustStringFormatter(
			"%{time:2006-01-02 15:04:05.000} %{level} - Memento - %{message}",
		)
	} else {
		format = logging.MustStringFormatter(
			"%{time:2006-01-02 15:04:05.000} %{level} - %{shortfile} - Memento - %{message}",
		)
	}

	if filename == "" {
		backend = logging.NewLogBackend(os.Stderr, "", 0)
	} else {
		fo, _ := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		backend = logging.NewLogBackend(fo, "", 0)
	}

	backendLeveled := logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(level, "")

	logging.SetBackend(backendLeveled)
	logging.SetFormatter(format)

	return log
}

func main() {
	var cfg *ini.File
	var backup bool
	var reload bool
	var err error

	s := options.NewOptions(SPEC)

	// Check if options isn't passed
	if len(os.Args[1:]) <= 0 {
		s.PrintUsageAndExit("No option specified")
	}

	opts := s.Parse(os.Args[1:])
	grace := ""

	// Print version and exit
	if opts.GetBool("version") {
		fmt.Println("Memento server " + VERSION)
		os.Exit(0)
	}

	// Print help and exit
	if opts.GetBool("help") {
		s.PrintUsageAndExit("Memento server " + VERSION)
	}

	// Check backup or restore operation (mandatory)
	if opts.GetBool("backup") && opts.GetBool("restore") {
		// If backup and restore options are passed in the same session, print help and exit
		s.PrintUsageAndExit("Cannot perform a backup and restore operations in the same session")
	}

	// Read grace (mandatory)
	if opts.GetBool("hour") {
		grace = "hour"
	} else if opts.GetBool("day") {
		grace = "day"
	} else if opts.GetBool("week") {
		grace = "week"
	} else if opts.GetBool("month") {
		grace = "month"
	} else {
		// If grace is not selected, print help and exit
		s.PrintUsageAndExit("No grace selected")
	}

	if opts.GetBool("reload-dataset") {
		reload = true
	} else {
		reload = false
	}

	if opts.GetBool("backup") {
		backup = true
		cfg, err = ini.Load([]byte{}, opts.Get("backup"))
	} else if opts.GetBool("restore") {
		backup = false
		cfg, err = ini.Load([]byte{}, opts.Get("backup"))
	}

	if err != nil {
		fmt.Println("Error about reading config file:", err)
		os.Exit(1)
	}

	loglevel, _ := logging.LogLevel(cfg.Section("general").Key("log_level").String())
	log := setlog(loglevel, cfg.Section("general").Key("log_file").String())

	log.Info("Started version " + VERSION)
	log.Debug("Grace selected: " + grace)

	if backup {
		repository := cfg.Section("general").Key("repository").String()
		check_structure(repository)

		server.Backup(log, cfg, grace, reload)
	} else {
		// TODO: add restore operation
	}

	log.Info("Ended version " + VERSION)
}
