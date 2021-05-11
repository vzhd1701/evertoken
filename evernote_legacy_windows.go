package main

import (
	"os"
	"path/filepath"
)

func legacyGetUsers() (Users []User) {
	confDir, err := os.UserHomeDir()
	panicFail(err)

	databasesPatterns := []string{
		filepath.Join(confDir, "Yinxiang Biji", "Databases", "*.exb"),
		filepath.Join(confDir, "Evernote", "Databases", "*.exb"),
	}

	for _, dbPattern := range databasesPatterns {
		dbFiles, err := filepath.Glob(dbPattern)
		panicFail(err)

		for _, f := range dbFiles {
			Users = append(Users, exbGetUser(f, "", -1))
		}
	}

	return
}
