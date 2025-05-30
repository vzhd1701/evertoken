package evernote

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/kirsle/configdir"
	"github.com/vzhd1701/evertoken/internal/myerrors"
	"github.com/vzhd1701/evertoken/internal/platform"
	"github.com/vzhd1701/evertoken/internal/types"
	"golang.org/x/text/encoding/charmap"
	_ "modernc.org/sqlite"
)

func NewGetUsers(newPath string, showRaw bool) (users []types.User) {
	confDir := getEvernoteDir(newPath)

	for _, user := range newListUsers(confDir) {
		storagePath := filepath.Join(confDir, "secure-storage", "authtoken_user_"+user.ID)

		userStoreDataRAW := getDecryptedSecureStorage(storagePath, user.ID)
		userStoreData := types.ParseRAWUserStoreData(userStoreDataRAW)

		user := types.User{
			Path:          storagePath,
			ID:            user.ID,
			UserName:      user.Username,
			Email:         user.Email,
			Token:         userStoreData.Token,
			UserStoreData: &userStoreData,
		}

		if showRaw {
			user.UserStoreDataRAW = userStoreDataRAW
		}

		users = append(users, user)
	}

	return
}

func NewGetSecureUsers(newSecurePath string, showRaw bool) (users []types.User) {
	userID := parseUserID(newSecurePath)

	userStoreDataRAW := getDecryptedSecureStorage(newSecurePath, userID)
	userStoreData := types.ParseRAWUserStoreData(userStoreDataRAW)

	user := types.User{
		Path:          newSecurePath,
		ID:            userID,
		Token:         userStoreData.Token,
		UserStoreData: &userStoreData,
	}

	if showRaw {
		user.UserStoreDataRAW = userStoreDataRAW
	}

	users = append(users, user)

	return
}

func parseUserID(filePath string) string {
	filename := filepath.Base(filePath)

	re := regexp.MustCompile(`^authtoken_user_(\d+)$`)

	match := re.FindStringSubmatch(filename)

	if len(match) != 2 {
		myerrors.ExpectedFail(fmt.Errorf("secure storage file '%s' does not match the expected format 'authtoken_user_<userID>'", filePath))
	}

	userID := match[1]
	_, err := strconv.Atoi(userID)
	if err != nil {
		myerrors.ExpectedFail(fmt.Errorf("parsed userID '%s' is not a valid integer: %w", userID, err))
	}

	return userID
}

func getDecryptedSecureStorage(storagePath string, userID string) string {
	encData, iv := newGetSecureStorageData(storagePath)

	key := newGetSecureStorageKey(userID)

	decryptedData := platform.AESDecrypt(encData, key, iv)

	return string(decryptedData)
}

func getEvernoteDir(customPath string) string {
	var userDir string

	if customPath == "" {
		userDir = configdir.LocalConfig("Evernote")
	} else {
		userDir = customPath
	}

	myerrors.FailIfNotAccessible(userDir, "Evernote user config directory")

	return userDir
}

func newListUsers(confDir string) (users []types.UserData) {
	dbFile := filepath.Join(confDir, "conduit-storage", "https%3A%2F%2Fwww.evernote.com", "_ConduitMultiUserDB.sql")

	myerrors.FailIfNotAccessible(dbFile, "user database file")

	db, err := sql.Open("sqlite", dbFile)
	myerrors.PanicFail(err)

	rows, err := db.Query("SELECT Tkey, TValue FROM MultiUsers")
	myerrors.PanicFail(err)

	defer func(rows *sql.Rows) {
		err := rows.Close()
		myerrors.PanicFail(err)
	}(rows)

	for rows.Next() {
		var userID string
		var userInfoS string

		err = rows.Scan(&userID, &userInfoS)
		myerrors.PanicFail(err)

		var userInfo map[string]interface{}
		err = json.Unmarshal([]byte(userInfoS), &userInfo)
		myerrors.PanicFail(err)

		users = append(users, types.UserData{
			ID:       strings.Split(userID, ":")[1],
			Username: userInfo["username"].(string),
			Email:    userInfo["email"].(string),
		})
	}

	return users
}

func newGetSecureStorageData(storagePath string) ([]byte, []byte) {
	myerrors.FailIfNotAccessible(storagePath, "user secure storage file")

	dat, err := os.ReadFile(storagePath)
	myerrors.PanicFail(err)

	var storageData map[string]interface{}
	err = json.Unmarshal(dat, &storageData)
	myerrors.PanicFail(err)

	// authtoken_user encrypted data saved as js "binary string" by Evernote
	// without properly encoding it to base64
	// emulating this behaviour by encoding string with ISO8859_1
	encryptedData, err := charmap.ISO8859_1.NewEncoder().String(storageData["encrypted"].(string))
	myerrors.PanicFail(err)

	iv, err := base64.StdEncoding.DecodeString(storageData["iv"].(string))
	myerrors.PanicFail(err)

	return []byte(encryptedData), iv
}

func newGetSecureStorageKey(userId string) []byte {
	var keyData string

	const keyPrefix = "enote-encr-key"

	service := "Evernote"
	accountID := "AuthToken:User:" + userId

	keyData = string(platform.GetSecureStorageData(service, accountID))

	key, err := base64.StdEncoding.DecodeString(strings.Replace(keyData, keyPrefix, "", 1))
	myerrors.PanicFail(err)

	return key
}
