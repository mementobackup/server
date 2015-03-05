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

var SECT_RESERVED = []string{"default", "general", "database", "dataset"}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func getdataset(grace int) int {
	return -1
}

func Sync(cfg *conf.ConfigFile, grace string) {
	var dataset, datasets int
	var sections []string

	sections = cfg.GetSections()
	datasets, _ = cfg.GetInt("dataset", grace)
	dataset = getdataset(datasets)
	fmt.Println("dataset: ", dataset)

	for _, section := range sections {
		if !contains(SECT_RESERVED, section) {
			fmt.Println(section) // TODO: add code for process a section
		}
	}
}
