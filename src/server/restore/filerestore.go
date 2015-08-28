/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package restore

import (
	"bitbucket.org/ebianchi/memento-common/common"
	"github.com/go-ini/ini"
	"github.com/op/go-logging"
)

func Filerestore(log *logging.Logger, section *common.Section, cfg *ini.File) {
	var cmd common.JSONMessage

	cmd.Context = "file"
	cmd.Command.Name = "put"
	cmd.Command.ACL = cfg.Section(section.Name).Key("acl").MustBool()

	if cfg.Section(section.Name).Key("path").String() == "" {
		// TODO: write full section restore
	} else {
		// TODO: write file restore
	}
}
