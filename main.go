package main

import (
	"fmt"
	"log"

	"github.com/integrii/flaggy"
)

var version = "0.0.1"

func init() {
	flaggy.SetName("evertoken")
	flaggy.SetDescription("Extract authentication token from Evernote.")
	flaggy.SetVersion(version)

	flaggy.DefaultParser.AdditionalHelpPrepend = "https://github.com/vzhd1701/evertoken"
}

func main() {
	var newPath, exbPath, exbPass string
	exbBruteStart := int64(-1)

	subcmdNew := flaggy.NewSubcommand("new")
	subcmdNew.Description = "Extract token from new Evernote app."

	subcmdNew.String(&newPath, "u", "user-dir", "Path to Evernote user config directory. (Optional, use only if you changed it)")

	subcmdLegacy := flaggy.NewSubcommand("legacy")
	subcmdLegacy.Description = "Extract token from legacy Evernote app."

	subcmdLegacyEXB := flaggy.NewSubcommand("legacy-exb")
	subcmdLegacyEXB.Description = "Extract token from EXB database file."

	subcmdLegacyEXB.AddPositionalValue(&exbPath, "exb", 1, true, "EXB database file path.")
	subcmdLegacyEXB.String(&exbPass, "p", "password", "Password to decrypt token data, numeric volume serial.")
	subcmdLegacyEXB.Int64(&exbBruteStart, "b", "brute", "Brute force password start number, use either this or password option.")

	flaggy.AttachSubcommand(subcmdNew, 1)
	flaggy.AttachSubcommand(subcmdLegacy, 1)
	flaggy.AttachSubcommand(subcmdLegacyEXB, 1)

	flaggy.Parse()

	var Users []User

	switch {
	case subcmdNew.Used:
		Users = newGetUsers(newPath)
	case subcmdLegacy.Used:
		Users = legacyGetUsers()
	case subcmdLegacyEXB.Used:
		Users = append(Users, exbGetUser(exbPath, exbPass, exbBruteStart))
	default:
		flaggy.ShowHelpAndExit("")
	}

	if len(Users) == 0 {
		fmt.Println("No users found! Make sure Evernote is installed, and there are logged-in users.")
		return
	}

	for _, user := range Users {
		printUser(user)
	}
}

func panicFail(err error) {
	if err != nil {
		log.Panic(fmt.Errorf("[ERROR] %w", err))
	}
}

func expectedFail(err error) {
	if err != nil {
		log.Fatal(fmt.Errorf("[ERROR] %w", err))
	}
}
