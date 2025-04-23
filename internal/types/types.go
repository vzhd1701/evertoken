package types

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-viper/mapstructure/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/vzhd1701/evertoken/internal/myerrors"
)

type User struct {
	Path             string
	ID               string
	UserName         string
	Email            string
	Token            string
	UserStoreDataRAW string
	UserStoreData    *UserStoreData
}

type UserData struct {
	ID       string
	Username string
	Email    string
}

type UserStoreData struct {
	Token        string         `mapstructure:"t,omitempty"`
	UserID       string         `mapstructure:"u,omitempty"`
	Shard        string         `mapstructure:"sh,omitempty"`
	Host         string         `mapstructure:"h,omitempty"`
	RefreshToken string         `mapstructure:"nrt,omitempty"`
	AccessToken  string         `mapstructure:"j,omitempty"`
	ClientID     string         `mapstructure:"nci,omitempty"`
	URLAccounts  string         `mapstructure:"nau,omitempty"`
	URLRedirect  string         `mapstructure:"nru,omitempty"`
	Unknown      map[string]any `mapstructure:",remain"`
}

func (user User) PrintDetails() {
	fmt.Printf("%v\n========================\n", user.Path)

	printIfNotEmpty("User ID", user.ID)
	printIfNotEmpty("Username", user.UserName)
	printIfNotEmpty("Email", user.Email)

	if user.Token != "" {
		printIfNotEmpty("Token", user.Token)
		printIfNotEmpty("Token EXP", getTokenExpiration(user.Token))
	} else {
		printIfNotEmpty("Token", "NOT FOUND")
	}

	if user.UserStoreData != nil {
		userStoreData := *user.UserStoreData

		if userStoreData.RefreshToken != "" {
			printIfNotEmpty("Refresh Token (JWT)", userStoreData.RefreshToken)
			printIfNotEmpty("Refresh Token (JWT) EXP", getJWTTokenExpiration(userStoreData.RefreshToken))
		}

		if userStoreData.AccessToken != "" {
			printIfNotEmpty("Access Token (JWT)", userStoreData.AccessToken)
			printIfNotEmpty("Access Token (JWT) EXP", getJWTTokenExpiration(userStoreData.AccessToken))
		}

		printIfNotEmpty("Shard", userStoreData.Shard)
		printIfNotEmpty("Host", userStoreData.Host)
		printIfNotEmpty("Client ID", userStoreData.ClientID)
		printIfNotEmpty("Accounts URL", userStoreData.URLAccounts)
		printIfNotEmpty("Redirect URL", userStoreData.URLRedirect)

		if len(userStoreData.Unknown) > 0 {
			printIfNotEmpty("Unknown Fields", fmt.Sprintf("%+v", userStoreData.Unknown))
		}
	}

	if user.UserStoreDataRAW != "" {
		fmt.Printf("\n%v\n========================\n", fmt.Sprintf("%s [RAW]", user.Path))
		fmt.Print(user.UserStoreDataRAW)
	}
}

func ParseRAWUserStoreData(data string) UserStoreData {
	storagedataRaw, err := base64.StdEncoding.DecodeString(data)
	myerrors.PanicFail(err)

	return parseUserStoreData(storagedataRaw)
}

func parseUserStoreData(data []byte) UserStoreData {
	var rawData map[string]any
	var resultData UserStoreData

	err := json.Unmarshal(data, &rawData)
	myerrors.PanicFail(err)

	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &resultData,
	}

	decoder, err := mapstructure.NewDecoder(config)
	myerrors.PanicFail(err)

	err = decoder.Decode(rawData)
	myerrors.PanicFail(err)

	return resultData
}

func printIfNotEmpty(key string, value string) {
	if value != "" {
		fmt.Printf("%-25s%v\n", key, value)
	}
}

func getTokenExpiration(token string) string {
	tokenMap := parseToken(token)

	expirationTimeRaw, ok := tokenMap["E"]

	if !ok {
		return "Error (bad token format - no E component)"
	}

	expirationTimeInt, err := strconv.ParseInt(expirationTimeRaw, 16, 64)
	myerrors.PanicFail(err)

	expirationTime := time.Unix(expirationTimeInt/1000, 0)

	return fmt.Sprintf("%s [%s]", expirationTime.String(), humanize.Time(expirationTime))
}

func parseToken(token string) map[string]string {
	result := make(map[string]string)

	pairs := strings.Split(token, ":")

	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)

		if len(kv) == 2 {
			key := kv[0]
			value := kv[1]
			result[key] = value
		}
	}

	return result
}

func getJWTTokenExpiration(token string) string {
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return fmt.Sprintf("Error (%s)", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)

	if !ok {
		return "Error (cannot parse claims)"
	}

	expirationTimeRaw, ok := claims["exp"]

	if !ok {
		return "Error (exp not found in claims)"
	}

	expirationTime := time.Unix(int64(expirationTimeRaw.(float64)), 0)

	return fmt.Sprintf("%s [%s]", expirationTime.String(), humanize.Time(expirationTime))
}
