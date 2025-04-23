package evernote

import (
	"bytes"
	"crypto/sha1"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/vzhd1701/evertoken/internal/myerrors"
	"github.com/vzhd1701/evertoken/internal/platform"
	"github.com/vzhd1701/evertoken/internal/types"
	"golang.org/x/crypto/pbkdf2"
)

var tokenLocations = map[string][4]int{
	"evernote":       {378, 155, 4, 0},
	"enscript":       {378, 155, 4, 1},
	"evernote_china": {471, 220, 4, 0},
	"enscript_china": {471, 220, 4, 1},
}

var userLocations = map[string][4]int{
	"evernote":       {378, 289, 4, 15},
	"enscript":       {378, 289, 4, 8},
	"evernote_china": {471, 372, 4, 15},
	"enscript_china": {471, 372, 4, 8},
}

func EXBGetUser(dbFile string, password string, bruteForceStart int64) types.User {
	if _, err := os.Stat(dbFile); errors.Is(err, os.ErrNotExist) {
		err = errors.New("database file does not exist")
		myerrors.ExpectedFail(err)
	}

	db, err := sql.Open("sqlite", dbFile)
	myerrors.PanicFail(err)

	token, err := exbGetToken(db, password, bruteForceStart)
	myerrors.ExpectedFail(err)

	user, err := exbGetUserData(db)
	myerrors.ExpectedFail(err)

	return types.User{
		Path:     dbFile,
		ID:       user.ID,
		UserName: user.Username,
		Email:    user.Email,
		Token:    token,
	}
}

func exbGetUserData(db *sql.DB) (user types.UserData, err error) {
	userData, err := exbGetKnownData(db, userLocations)

	if err != nil {
		userData, err = exbBlindSearchUser(db)

		if err != nil {
			return
		}
	}

	user = exbReadUserData(userData)

	return
}

func exbBlindSearchUser(db *sql.DB) (data []byte, err error) {
	for _, rowData := range exbBlindDataQuery(db) {
		if len(rowData) > 500 && bytes.HasPrefix(rowData, []byte{8, 0, 1}) {
			data = rowData
			return
		}
	}

	err = errors.New("user data not found on blind search")

	return
}

func exbGetToken(db *sql.DB, password string, bruteForceStart int64) (token string, err error) {
	tokenData, err := exbGetKnownData(db, tokenLocations)

	if err != nil {
		tokenData, err = exbBlindSearchToken(db)

		if err != nil {
			return
		}
	}

	if bruteForceStart != -1 {
		password, err = exbBruteForcePass(tokenData, bruteForceStart)
		myerrors.ExpectedFail(err)

		fmt.Println("\nPassword is:", password)
	}

	token, err = exbDecryptToken(tokenData, password)

	return
}

func exbBlindSearchToken(db *sql.DB) (data []byte, err error) {
	for _, rowData := range exbBlindDataQuery(db) {
		if len(rowData)%16 == 0 && len(rowData) > 200 && len(rowData) < 300 && platform.Shannon(rowData[:200]) >= 1000 {
			data = rowData
			return
		}
	}

	err = errors.New("token not found on blind search")

	return
}

func exbBlindDataQuery(db *sql.DB) (datas [][]byte) {
	rows, err := db.Query("SELECT data FROM attrs")
	myerrors.PanicFail(err)

	defer func(rows *sql.Rows) {
		err := rows.Close()
		myerrors.PanicFail(err)
	}(rows)

	for rows.Next() {
		var data []byte

		err = rows.Scan(&data)
		myerrors.PanicFail(err)

		datas = append(datas, data)
	}

	return
}

func exbGetKnownData(db *sql.DB, knownLocations map[string][4]int) (resultData []byte, err error) {
	for _, loc := range knownLocations {
		data, err := exbQueryData(db, loc)
		if err != nil {
			continue
		}

		resultData = data
	}

	if len(resultData) == 0 {
		err = errors.New("not found in known locations")
	}

	return
}

func exbQueryData(db *sql.DB, location [4]int) (data []byte, err error) {
	row := db.QueryRow("SELECT data FROM attrs WHERE uid=? and aid=? and afl=? and csn=?",
		location[0], location[1], location[2], location[3])

	err = row.Scan(&data)

	return
}

func exbReadUserData(data []byte) (user types.UserData) {
	buf := bytes.NewBuffer(data)

	buf.Next(3)
	user.ID = strconv.Itoa(int(binary.BigEndian.Uint32(buf.Next(4))))

	buf.Next(3)
	user.Username = exbReadString(buf)

	buf.Next(3)
	user.Email = exbReadString(buf)

	return
}

func exbReadString(buf *bytes.Buffer) string {
	strLen := int(binary.BigEndian.Uint32(buf.Next(4)))
	return string(buf.Next(strLen))
}

func exbDecryptToken(data []byte, password string) (token string, err error) {
	cleanData := exbUnscrambleTokenData(data)

	if password == "" {
		password = platform.GetDiskSerial()
	}

	key := exbMakeKey(password)

	iv := cleanData[:16]
	ciphertext := cleanData[16:]

	token = string(platform.AESDecrypt(ciphertext, key, iv))

	if !strings.HasPrefix(token, "S=s") {
		err = errors.New("failed to decrypt token data")
	}

	return
}

func exbUnscrambleTokenData(obfData []byte) []byte {
	var deobfData []byte

	for i, char := range obfData {
		curSize := byte(len(obfData) - i)
		newByte := char ^ ((curSize * (curSize ^ 0xFF)) & 0x7F)

		deobfData = append(deobfData, newByte)
	}

	data, err := hex.DecodeString(string(deobfData))
	myerrors.PanicFail(err)

	return data
}

func exbMakeKey(password string) []byte {
	const passwordBase = "{154BE163-907C-4188-9929-37F9A2D4EFD4}"
	const salt = "{91E71FB1-F55F-4fbc-B9FD-DC0A8C01C469}"

	passwordSeed := []byte(passwordBase + password)

	return pbkdf2.Key(passwordSeed, []byte(salt), 4096, 16, sha1.New)
}

func exbBruteWorker(ciphertext []byte, iv []byte, jobs <-chan int64, results chan<- string) {
	tokenPrefix := []byte("S=s")

	for j := range jobs {
		password := strconv.FormatInt(j, 10)

		key := exbMakeKey(password)

		tokenBytes := platform.AESDecrypt(ciphertext, key, iv)

		if !bytes.HasPrefix(tokenBytes, tokenPrefix) {
			results <- ""
		} else {
			results <- password
		}
	}
}

func exbBruteForcePass(encData []byte, start int64) (password string, err error) {
	const numJobsTotal int64 = 4294967295

	jobsBatch := int64(100000)

	bar := progressbar.Default(numJobsTotal)

	err = bar.Add64(start)
	myerrors.PanicFail(err)

	cleanData := exbUnscrambleTokenData(encData)

	iv := cleanData[:16]
	ciphertext := cleanData[16:]

	for i := start; i < numJobsTotal; {
		// last batch
		if numJobsTotal-i < jobsBatch {
			jobsBatch = numJobsTotal - i
		}

		jobs := make(chan int64, jobsBatch)
		results := make(chan string, jobsBatch)

		for w := 1; w <= 10; w++ {
			go exbBruteWorker(ciphertext, iv, jobs, results)
		}

		for j := i; j <= i+jobsBatch; j++ {
			jobs <- j
		}

		close(jobs)

		for a := int64(1); a <= jobsBatch; a++ {
			resPassword := <-results

			err := bar.Add64(1)
			myerrors.PanicFail(err)

			if resPassword != "" {
				password = resPassword
				return password, nil
			}
		}

		if i+jobsBatch > numJobsTotal {
			i = numJobsTotal - i
		} else {
			i += jobsBatch
		}
	}

	err = errors.New("password not found")

	return
}
