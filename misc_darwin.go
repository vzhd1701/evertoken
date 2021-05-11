package main

import (
	"errors"
	"os/exec"
)

// Stub for darwin
func getDiskSerial() string {
	return ""
}

func getSecureStorageData(service string, accountID string) (keyDataBin []byte) {
	var err error

	if accountID == "" {
		keyDataBin, err = exec.Command(
			"/usr/bin/security",
			"find-generic-password",
			"-s", service).CombinedOutput()
	} else {
		keyDataBin, err = exec.Command(
			"/usr/bin/security",
			"find-generic-password",
			"-s", service,
			"-wa", accountID).CombinedOutput()
	}

	if err != nil {
		switch err.Error() {
		case "exit status 44":
			err = errors.New("entry not found in secure storage")
		case "exit status 128":
			err = errors.New("secure storage access denied")
		default:
			panicFail(err)
		}
		expectedFail(err)
	}

	return keyDataBin
}
