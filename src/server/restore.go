/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package server

import (
	"github.com/go-ini/ini"
	"github.com/op/go-logging"
)

func Restore(log *logging.Logger, cfg *ini.File, grace string) {
	dataset := cfg.Section("general").Key("dataset").MustInt()

	for _, section := range cfg.Sections() {
		if !contains(SECT_RESERVED, section.Name()) {
			if section.Key("type").String() == "file" {
				restorefile(log, cfg, grace, dataset)
			}
		}
	}
}

func restorefile(log *logging.Logger, cfg *ini.File, grace, dataset string) {

}
