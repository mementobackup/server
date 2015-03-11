/*
 Copyright (C) 2015 Enrico Bianchi (enrico.bianchi@gmail.com)
 Project       Memento
 Description   A backup system
 License       GPL version 2 (see GPL.txt for details)
*/

package database

func Getdataset(db *DB, grace string) int {
	var result int

	var query = "SELECT actual FROM status WHERE grace = ?"

	err := db.Conn.QueryRow(query, grace).Scan(&result)
	if err != nil {
		// TODO: add log point
		return 0
	}
	return result
}
