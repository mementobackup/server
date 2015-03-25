/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package generic

import "github.com/go-ini/ini"

type Section struct {
	Section *ini.Section
	Grace   string
	Dataset int
}
