package main

import (
	"fmt"

	"github.com/integrii/flaggy"
	"github.com/vzhd1701/evertoken/internal/evernote"
	"github.com/vzhd1701/evertoken/internal/types"
)

var version = "0.1.0"

func init() {
	flaggy.SetName("evertoken")
	flaggy.SetDescription("Extract authentication token from Evernote.")
	flaggy.SetVersion(version)

	flaggy.DefaultParser.AdditionalHelpPrepend = "https://github.com/vzhd1701/evertoken"
}

func main() {
	var newPath, newSecurePath, exbPath, exbPass string
	var showRaw bool
	exbBruteStart := int64(-1)

	subcmdNew := flaggy.NewSubcommand("new")
	subcmdNew.Description = "Extract token from modern Evernote app."

	subcmdNew.String(&newPath, "u", "user-dir", "Path to Evernote user config directory. (Optional, use only if you moved it)")
	subcmdNew.Bool(&showRaw, "r", "show-raw", "Show raw decrypted secure storage data.")

	subcmdNewSS := flaggy.NewSubcommand("new-ss")
	subcmdNewSS.Description = "Extract token directly from Evernote's secure storage file authtoken_user_<userID>."

	subcmdNewSS.AddPositionalValue(&newSecurePath, "ss-file", 1, true, "Path to Evernote secure storage file authtoken_user_<userID>.")
	subcmdNewSS.Bool(&showRaw, "r", "show-raw", "Show raw decrypted secure storage data.")

	subcmdLegacy := flaggy.NewSubcommand("legacy")
	subcmdLegacy.Description = "Extract token from legacy Evernote app."

	subcmdLegacyEXB := flaggy.NewSubcommand("legacy-exb")
	subcmdLegacyEXB.Description = "Extract token from EXB database file."

	subcmdLegacyEXB.AddPositionalValue(&exbPath, "exb", 1, true, "EXB database file path.")
	subcmdLegacyEXB.String(&exbPass, "p", "password", "Password to decrypt token data, numeric volume serial.")
	subcmdLegacyEXB.Int64(&exbBruteStart, "b", "brute", "Brute force password start number, use either this or password option.")

	flaggy.AttachSubcommand(subcmdNew, 1)
	flaggy.AttachSubcommand(subcmdNewSS, 1)
	flaggy.AttachSubcommand(subcmdLegacy, 1)
	flaggy.AttachSubcommand(subcmdLegacyEXB, 1)

	flaggy.Parse()

	var Users []types.User

	switch {
	case subcmdNew.Used:
		Users = evernote.NewGetUsers(newPath, showRaw)
	case subcmdNewSS.Used:
		Users = evernote.NewGetSecureUsers(newSecurePath, showRaw)
	case subcmdLegacy.Used:
		Users = evernote.LegacyGetUsers()
	case subcmdLegacyEXB.Used:
		Users = append(Users, evernote.EXBGetUser(exbPath, exbPass, exbBruteStart))
	default:
		flaggy.ShowHelpAndExit("")
	}

	if len(Users) == 0 {
		fmt.Println("No users found! Make sure Evernote is installed, and there are logged-in users.")
		return
	}

	for _, user := range Users {
		user.PrintDetails()
	}
}
