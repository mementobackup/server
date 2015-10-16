/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package backup

import (
	"bitbucket.org/ebianchi/memento-common/common"
	"github.com/go-ini/ini"
	"github.com/op/go-logging"
	"net"
	"os"
	"path/filepath"
	"server/dataset"
	"server/network"
)

func fs_save_data(log *logging.Logger, cfg *ini.File, section *common.Section, data common.JSONFile, previous bool) {
	var item, source, dest, hash string
	var cmd common.JSONMessage
	var conn net.Conn
	var err error

	item = dataset.Convertpath(data.Os, data.Name)
	source = dataset.Path(cfg, section, true) + string(filepath.Separator) + item
	dest = dataset.Path(cfg, section, false) + string(filepath.Separator) + item

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
			if section.Compressed {
				source = source + ".compressed"
				dest = dest + ".compressed"
			}

			if err = os.Link(source, dest); err != nil {
				log.Error("Error when link file %s", data.Name)
				log.Debug("Trace: " + err.Error())
			}
		} else {
			cmd.Context = "file"
			cmd.Command.Name = "get"
			cmd.Command.Element = data

			conn, err = network.Getsocket(cfg.Section(section.Name))
			if err != nil {
				log.Error("Error when opening connection with section " + section.Name)
				log.Debug("Trace: " + err.Error())
				return
			}
			defer conn.Close()

			cmd.Send(conn)

			if hash, err = common.Receivefile(dest, conn); err != nil {
				log.Error("Error when receiving file " + data.Name)
				log.Debug("Trace: " + err.Error())
			}

			// TODO: check file's hash
			if hash == "" {
				log.Error("Hash for file " + dest + " mismatch")
				// TODO: remove file if hash mismatch
			} else {
				log.Debug("Hash for file " + dest + " is " + hash)

				if section.Compressed {
					dataset.Compressfile(log, dest)
				}
			}
		}
	}
}
