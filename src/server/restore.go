/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package server

import (
	"bitbucket.org/ebianchi/memento-common/common"
	"github.com/go-ini/ini"
	"github.com/op/go-logging"
	"restore"
)

func Restore(log *logging.Logger, cfg *ini.File, grace string) {
	dataset := cfg.Section("general").Key("dataset").MustInt()

	for _, section := range cfg.Sections() {
		if !contains(SECT_RESERVED, section.Name()) {
			if section.Key("type").String() == "file" {
				sect := common.Section{
					Name:       section.Name(),
					Grace:      grace,
					Dataset:    dataset,
					Compressed: section.Key("compress").MustBool(),
				}
				filerestore(log, cfg, &sect)
			}
		}
	}
}

func filerestore(log *logging.Logger, cfg *ini.File, section *common.Section) {
	// Execute pre_command
	exec_command(log, cfg.Section(section.Name), "pre_command")

	// Restore!
	restore.Filerestore(log, section, cfg)

	// Execute post_command
	exec_command(log, cfg.Section(section.Name), "post_command")
}
