package myerrors

import (
	"errors"
	"fmt"
	"log"
	"os"
)

func PanicFail(err error) {
	if err != nil {
		log.Panic(fmt.Errorf("[ERROR] %w", err))
	}
}

func ExpectedFail(err error) {
	if err != nil {
		log.Fatal(fmt.Errorf("[ERROR] %w", err))
	}
}

func FailIfNotAccessible(path string, pathDescription string) {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		errorMessage := fmt.Sprintf("%s '%s' does not exist", pathDescription, path)
		err = errors.New(errorMessage)
		ExpectedFail(err)
	} else if os.IsPermission(err) {
		errorMessage := fmt.Sprintf("cannot access %s '%s', try running as admin", pathDescription, path)
		err = errors.New(errorMessage)
		ExpectedFail(err)
	} else if err != nil {
		PanicFail(err)
	}
}
