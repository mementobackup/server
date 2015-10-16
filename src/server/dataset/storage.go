/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package dataset

import (
	"bitbucket.org/ebianchi/memento-common/common"
	"compress/gzip"
	"github.com/go-ini/ini"
	"github.com/op/go-logging"
	"io"
	"os"
	"path/filepath"
	"server/database"
	"strconv"
	"strings"
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

func Compressfile(log *logging.Logger, filename string) {
	var filesource, filedest *os.File
	var err error

	if filesource, err = os.Open(strings.TrimSpace(filename)); err != nil {
		log.Error("Cannot open file " + filename + " for compression")
		log.Debug("Trace: " + err.Error())
		// FIXME: remove file if is not able to compress it?
		return
	}
	defer filesource.Close()

	if filedest, err = os.Create(strings.TrimSpace(filename) + ".compressed"); err != nil {
		log.Error("Compress for file " + filename + " failed")
		log.Debug("Trace: " + err.Error())
		// FIXME: remove file if is not able to compress it?
		return
	}
	defer filedest.Close()

	compress := gzip.NewWriter(filedest)
	defer compress.Close()

	if _, err = io.Copy(compress, filesource); err != nil {
		log.Error("Compress for file " + filename + " failed")
		log.Debug("Trace: " + err.Error())
		// FIXME: remove file if is not able to compress it?
		return
	}
	os.Remove(filename)
}

func Path(cfg *ini.File, section *common.Section, previous bool) string {
	var dataset int

	if previous {
		if section.Dataset == 1 {
			dataset = cfg.Section("dataset").Key(section.Grace).MustInt()
		} else {
			dataset = section.Dataset - 1
		}
	} else {
		dataset = section.Dataset
	}

	destination := strings.Join([]string{
		cfg.Section("general").Key("repository").String(),
		section.Grace,
		strconv.Itoa(dataset),
		section.Name,
	}, string(filepath.Separator))

	return destination
}
