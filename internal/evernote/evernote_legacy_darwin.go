package evernote

import (
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"

	"github.com/vzhd1701/evertoken/internal/myerrors"
	"github.com/vzhd1701/evertoken/internal/platform"
	"github.com/vzhd1701/evertoken/internal/types"
	"howett.net/plist"
)

func LegacyGetUsers() (Users []types.User) {
	for _, accountID := range legacyMacGetAccounts() {
		Users = append(Users, legacyMacGetUser(accountID))
	}

	return
}

func legacyMacGetAccounts() (accounts []string) {
	evernoteAccounts := platform.GetSecureStorageData("Evernote", "")

	r, _ := regexp.Compile("\"acct\"<blob>=\"([0-9]+/Evernote/smd)\"")

	for _, userMatch := range r.FindAllStringSubmatch(string(evernoteAccounts), -1) {
		accountID := userMatch[1]
		accounts = append(accounts, accountID)
	}

	return
}

func legacyMacGetUser(accountID string) types.User {
	accountDataHex := platform.GetSecureStorageData("Evernote", accountID)

	accountDataBin, err := hex.DecodeString(strings.TrimSpace(string(accountDataHex)))
	myerrors.PanicFail(err)

	accountData := make(map[string]interface{})

	_, err = plist.Unmarshal(accountDataBin, &accountData)
	myerrors.PanicFail(err)

	return types.User{
		Path:     accountID,
		ID:       strconv.FormatUint(accountData["$objects"].([]interface{})[6].(uint64), 10),
		UserName: accountData["$objects"].([]interface{})[7].(string),
		Email:    accountData["$objects"].([]interface{})[8].(string),
		Token:    accountData["$objects"].([]interface{})[3].(string),
	}
}
