/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package syncing

import (
	"bitbucket.org/ebianchi/memento-common/common"
	"github.com/go-ini/ini"
	"github.com/op/go-logging"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func fs_compute_dest(path string, cfg *ini.File, section *common.Section, previous bool) string {
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

func fs_save_data(log *logging.Logger, cfg *ini.File, section *common.Section, data common.JSONFile, previous bool) {
	var item string
	var err error

	if data.Os == "windows" {
		item = strings.Replace(data.Name, "\\", "/", -1)
	} else {
		item = data.Name
	}

	dest := fs_compute_dest(data.Name, cfg, section, false) + string(filepath.Separator) + item
	log.Debug("Save item: " + dest)

	switch data.Type {
	case "directory":
		os.MkdirAll(dest, 0755)
	case "symlink":
		if err := os.Symlink(data.Link, dest); err != nil {
			log.Error("Error when creating symlink for file %s", data.Name)
			log.Debug("Trace: %s", err.Error())
		}
	case "file":
		if previous {
			source := fs_compute_dest(data.Name, cfg, section, true) + string(filepath.Separator) + item

			if section.Compressed {
				err = os.Link(source + ".compressed", dest + ".compressed")
			} else {
				err = os.Link(source, dest)
			}

			if err != nil {
				log.Error("Error when link file %s", data.Name)
				log.Debug("Trace: " + err.Error() )
			}
		} else {
			// TODO: write code for downloading and saving file
		}
	}
}
