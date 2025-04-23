package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/vzhd1701/evertoken/internal/myerrors"
)

type User struct {
	Path     string
	ID       string
	UserName string
	Email    string
	Token    string
}

type UserData struct {
	ID       string
	Username string
	Email    string
}

func (user User) PrintDetails() {
	fmt.Printf("%v\n========================\n", user.Path)

	if user.ID != "" {
		fmt.Printf("%-10s%v\n", "User ID", user.ID)
	}
	if user.UserName != "" {
		fmt.Printf("%-10s%v\n", "Username", user.UserName)
	}
	if user.Email != "" {
		fmt.Printf("%-10s%v\n", "Email", user.Email)
	}
	if user.Token != "" {
		fmt.Printf("%-10s%v\n", "Token", user.Token)
		fmt.Printf("%-10s%v\n", "Token Exp", getTokenExpiration(user.Token))
	}
}

func getTokenExpiration(token string) string {
	tokenParts := strings.Split(token, ":")

	expirationTime, err := strconv.ParseInt(tokenParts[2][2:], 16, 64)
	myerrors.PanicFail(err)

	tm := time.Unix(expirationTime/1000, 0)

	return fmt.Sprintf("%s [%s]", tm.String(), humanize.Time(tm))
}
