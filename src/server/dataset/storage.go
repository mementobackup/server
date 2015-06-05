/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package dataset

import (
	"github.com/go-ini/ini"
	"github.com/op/go-logging"
	"os"
	"path/filepath"
	"server/database"
	"strconv"
)

func Deldataset(log *logging.Logger, cfg *ini.File, section, grace string, dataset int) {
	var db database.DB

	log.Debug("About to delete dataset " + strconv.Itoa(dataset) + " for section " + section + " and grace " + grace)

	db.Open(log, cfg)
	defer db.Close()

	err := database.Deldataset(log, &db, section, grace, dataset)
	if err != nil {
		log.Error(err.Error())
	}

	repository := cfg.Section("general").Key("repository").String() +
		string(filepath.Separator) +
		grace +
		string(filepath.Separator) +
		strconv.Itoa(dataset)

	if section != "" {
		repository = repository + string(filepath.Separator) + section
	}

	os.RemoveAll(repository)

	log.Debug("Dataset deleted")
}