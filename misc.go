package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

type User struct {
	path     string
	id       string
	username string
	email    string
	token    string
}

type UserData struct {
	ID       string
	Username string
	Email    string
}

func printUser(user User) {
	fmt.Printf("%v\n========================\n", user.path)

	if user.id != "" {
		fmt.Printf("%-10s%v\n", "User ID", user.id)
	}
	if user.username != "" {
		fmt.Printf("%-10s%v\n", "Username", user.username)
	}
	if user.email != "" {
		fmt.Printf("%-10s%v\n", "Email", user.email)
	}
	if user.token != "" {
		fmt.Printf("%-10s%v\n", "Token", user.token)
		fmt.Printf("%-10s%v\n", "Token Exp", getTokenExpiration(user.token))
	}
}

func getTokenExpiration(token string) string {
	tokenParts := strings.Split(token, ":")

	expirationTime, err := strconv.ParseInt(tokenParts[2][2:], 16, 64)
	panicFail(err)

	tm := time.Unix(expirationTime/1000, 0)

	return fmt.Sprintf("%s [%s]", tm.String(), humanize.Time(tm))
}

func failIfNotAccessible(path string, pathDescription string) {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		errorMessage := fmt.Sprintf("%s '%s' does not exist", pathDescription, path)
		err = errors.New(errorMessage)
		expectedFail(err)
	} else if os.IsPermission(err) {
		errorMessage := fmt.Sprintf("cannot access %s '%s', try running as admin", pathDescription, path)
		err = errors.New(errorMessage)
		expectedFail(err)
	} else if err != nil {
		panicFail(err)
	}
}
