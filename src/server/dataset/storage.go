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
	"strings"
	"compress/gzip"
	"io"
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