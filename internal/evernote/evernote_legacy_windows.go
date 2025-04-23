package evernote

import (
	"os"
	"path/filepath"

	"github.com/vzhd1701/evertoken/internal/myerrors"
	"github.com/vzhd1701/evertoken/internal/types"
)

func LegacyGetUsers() (Users []types.User) {
	confDir, err := os.UserHomeDir()
	myerrors.PanicFail(err)

	databasesPatterns := []string{
		filepath.Join(confDir, "Yinxiang Biji", "Databases", "*.exb"),
		filepath.Join(confDir, "Evernote", "Databases", "*.exb"),
	}

	for _, dbPattern := range databasesPatterns {
		dbFiles, err := filepath.Glob(dbPattern)
		myerrors.PanicFail(err)

		for _, f := range dbFiles {
			Users = append(Users, EXBGetUser(f, "", -1))
		}
	}

	return
}
