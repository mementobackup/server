/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package server

import (
	"code.google.com/p/goconf/conf"
	"fmt"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Sync(cfg *conf.ConfigFile, grace string) {
	sections := cfg.GetSections()

	for _, section := range sections {
		if !contains([]string{"default", "general", "database", "dataset"}, section) {
			fmt.Println(section) // TODO: add code for process a section
		}
	}
}
