/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package main

import (
	"code.google.com/p/goconf/conf"
	"fmt"
	"github.com/gaal/go-options/options"
	"os"
	"path/filepath"
	"server"
)

const VERSION = "1.0"

const SPEC = `
Memento - A backup system
Usage: mserver [OPTIONS]
--
h,help                Print this help
v,version             Print version
c,cfg=                Open the configuration file
H,hour                Hourly backup
D,day                 Daily backup
W,week                Weekly backup
M,month               Monthly backup
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

func main() {
	s := options.NewOptions(SPEC)

	// Check if options isn't passed
	if len(os.Args[1:]) <= 0 {
		s.PrintUsageAndExit("No option specified")
	}

	opts := s.Parse(os.Args[1:])
	grace := ""

	// Read Ini file (mandatory)
	if !opts.GetBool("cfg") {
		// If configuration file is not passed, print help and exit
		s.PrintUsageAndExit("No configuration file passed")
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

	// Print version and exit
	if opts.GetBool("version") {
		fmt.Println("Memento server " + VERSION)
		os.Exit(0)
	}

	// Print help and exit
	if opts.GetBool("help") {
		s.PrintUsageAndExit("Memento server " + VERSION)
	}

	cfg, err := conf.ReadConfigFile(opts.Get("cfg"))
	if err != nil {
		fmt.Println("Error about reading config file:", err)
		os.Exit(1)
	}

	repository, _ := cfg.GetString("general", "repository")
	check_structure(repository)

	server.Sync(cfg, grace)
}
