package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
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

	return tm.String()
}
