/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package server

import (
	"bitbucket.org/ebianchi/memento-common/common"
)

func fs_metadata() {
	// TODO: write code for getting file's metadata from client
}


func syncfile(section *Section) {

	// Execute pre_command
	if section.Section.Key("pre_command").String() != "" {
		common.ExecuteCMD(section.Section.Key("pre_command").String())
	}

	fs_metadata()

	// Execute post_command
	if section.Section.Key("post_command").String() != "" {
		common.ExecuteCMD(section.Section.Key("post_command").String())
	}
}
