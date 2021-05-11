package main

import (
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"

	"howett.net/plist"
)

func legacyGetUsers() (Users []User) {
	for _, accountID := range legacyMacGetAccounts() {
		Users = append(Users, legacyMacGetUser(accountID))
	}

	return
}

func legacyMacGetAccounts() (accounts []string) {
	evernoteAccounts := getSecureStorageData("Evernote", "")

	r, _ := regexp.Compile("\"acct\"<blob>=\"([0-9]+/Evernote/smd)\"")

	for _, userMatch := range r.FindAllStringSubmatch(string(evernoteAccounts), -1) {
		accountID := userMatch[1]
		accounts = append(accounts, accountID)
	}

	return
}

func legacyMacGetUser(accountID string) User {
	accountDataHex := getSecureStorageData("Evernote", accountID)

	accountDataBin, err := hex.DecodeString(strings.TrimSpace(string(accountDataHex)))
	panicFail(err)

	accountData := make(map[string]interface{})

	_, err = plist.Unmarshal(accountDataBin, &accountData)
	panicFail(err)

	return User{
		path:     accountID,
		id:       strconv.FormatUint(accountData["$objects"].([]interface{})[6].(uint64), 10),
		username: accountData["$objects"].([]interface{})[7].(string),
		email:    accountData["$objects"].([]interface{})[8].(string),
		token:    accountData["$objects"].([]interface{})[3].(string),
	}
}
