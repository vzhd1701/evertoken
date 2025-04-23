package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/kirsle/configdir"
	"golang.org/x/text/encoding/charmap"
	_ "modernc.org/sqlite"
)

func getEvernoteDir(customPath string) string {
	var userDir string

	if customPath == "" {
		userDir = configdir.LocalConfig("Evernote")
	} else {
		userDir = customPath
	}

	failIfNotAccessible(userDir, "Evernote user config directory")

	return userDir
}

func newGetUsers(newPath string) (users []User) {
	confDir := getEvernoteDir(newPath)

	for _, user := range newListUsers(confDir) {
		storagePath := filepath.Join(confDir, "secure-storage", "authtoken_user_"+user.ID)

		encData, iv := newGetSecureStorageData(storagePath)

		key := newGetSecureStorageKey(user.ID)

		decryptedData := aesDecrypt(encData, key, iv)

		token := newGetToken(decryptedData)

		users = append(users, User{
			path:     storagePath,
			id:       user.ID,
			username: user.Username,
			email:    user.Email,
			token:    token,
		})
	}

	return
}

func newListUsers(confDir string) (users []UserData) {
	dbFile := filepath.Join(confDir, "conduit-storage", "https%3A%2F%2Fwww.evernote.com", "_ConduitMultiUserDB.sql")

	failIfNotAccessible(dbFile, "user database file")

	db, err := sql.Open("sqlite", dbFile)
	panicFail(err)

	rows, err := db.Query("SELECT Tkey, TValue FROM MultiUsers")
	panicFail(err)

	defer func(rows *sql.Rows) {
		err := rows.Close()
		panicFail(err)
	}(rows)

	for rows.Next() {
		var userID string
		var userInfoS string

		err = rows.Scan(&userID, &userInfoS)
		panicFail(err)

		var userInfo map[string]interface{}
		err = json.Unmarshal([]byte(userInfoS), &userInfo)
		panicFail(err)

		users = append(users, UserData{
			ID:       strings.Split(userID, ":")[1],
			Username: userInfo["username"].(string),
			Email:    userInfo["email"].(string),
		})
	}

	return users
}

func newGetSecureStorageData(storagePath string) ([]byte, []byte) {
	failIfNotAccessible(storagePath, "user secure storage file")

	dat, err := os.ReadFile(storagePath)
	panicFail(err)

	var storageData map[string]interface{}
	err = json.Unmarshal(dat, &storageData)
	panicFail(err)

	// authtoken_user encrypted data saved as js "binary string" by Evernote
	// without properly encoding it to base64
	// emulating this behaviour by encoding string with ISO8859_1
	encryptedData, err := charmap.ISO8859_1.NewEncoder().String(storageData["encrypted"].(string))
	panicFail(err)

	iv, err := base64.StdEncoding.DecodeString(storageData["iv"].(string))
	panicFail(err)

	return []byte(encryptedData), iv
}

func newGetSecureStorageKey(userId string) []byte {
	var keyData string

	const keyPrefix = "enote-encr-key"

	service := "Evernote"
	accountID := "AuthToken:User:" + userId

	keyData = string(getSecureStorageData(service, accountID))

	key, err := base64.StdEncoding.DecodeString(strings.Replace(keyData, keyPrefix, "", 1))
	panicFail(err)

	return key
}

func newGetToken(storagedataBytes []byte) string {
	storagedataRaw, err := base64.StdEncoding.DecodeString(string(storagedataBytes))
	panicFail(err)

	var storageData map[string]interface{}
	err = json.Unmarshal(storagedataRaw, &storageData)
	panicFail(err)

	return storageData["t"].(string)
}
