package bot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)
const (
	StartCommand int = iota
	MeCommand
	ReKeyCommand
	BlockCommand
)
var Token string
var Secret string
var startCommand = "/start"
var startCommandLen = len(startCommand)

var meCommand = "/me"
var meCommandLen = len(meCommand)

var reKeyCommand = "/rekey"
var reKeyCommandLen = len(reKeyCommand)

var blockCommand = "/block"
var blockCommandLen = len(blockCommand)

func ParseCommand(text string) (int, string) {
	startCommandRegex, _ := regexp.Compile("^/start")
	meCommandRegex, _ := regexp.Compile("^/me")
	reKeyCommandRegex, _ := regexp.Compile("^/rekey")
	blockCommandRegex, _ := regexp.Compile("^/block")
	if startCommandRegex.MatchString(text) {
		return StartCommand, text[startCommandLen:]
	} else if meCommandRegex.MatchString(text) {
		return MeCommand, text[meCommandLen:]
	} else if reKeyCommandRegex.MatchString(text) {
		return ReKeyCommand, text[reKeyCommandLen:]
	} else if blockCommandRegex.MatchString(text) {
		return BlockCommand, strings.TrimSpace(text[blockCommandLen:])
	}
	return -1, ""
}

func SendMessage(chatID int, text string) {
	const telegramApiBaseUrl string = "https://api.telegram.org/bot"
	const telegramApiSendMessage string = "/sendMessage"

	var telegramApi = telegramApiBaseUrl + Token + telegramApiSendMessage
	response, err := http.PostForm(telegramApi, url.Values{
		"chat_id": {strconv.Itoa(chatID)},
		"text": {text},
	})
	if err != nil{
		fmt.Println(err)
	}
	b, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(b))
	if err != nil {
		return
	}
	defer response.Body.Close()
}

func SendBotCommandMessage(chatID int, text string, command string, buttonText string) {
	const telegramApiBaseUrl string = "https://api.telegram.org/bot"
	const telegramApiSendMessage string = "/sendMessage"

	var telegramApi = telegramApiBaseUrl + Token + telegramApiSendMessage
	var message = map[string][][]map[string]string{
		"inline_keyboard": {
			{
				{"text": buttonText, "callback_data": command},
			},
		},
	}
	b, _ := json.Marshal(&message)


	response, err := http.PostForm(telegramApi, url.Values{
		"chat_id": {strconv.Itoa(chatID)},
		"text": {text},
		"reply_markup": {
			string(b),
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	bo, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(bo))
	defer response.Body.Close()
}




func GenerateUUID() string {
	uid := uuid.New().String()
	return strings.ReplaceAll(uid, "-", "")
}

func AuthenticateRequestSecret(sec string) (string, error){
	Token, err := jwt.Parse(sec, func(Token *jwt.Token) (interface{}, error) {
		if _, ok := Token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", Token.Header["alg"])
		}
		return []byte(Secret), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := Token.Claims.(jwt.MapClaims); ok && Token.Valid {
		fmt.Println(claims["uid"].(string))
		return claims["uid"].(string), nil
	} else {
		return "", errors.New("error: Token invalid")
	}
}

func RespondCallbackQuery(queryID, text string) {
	const telegramApiBaseUrl string = "https://api.telegram.org/bot"
	const path string = "/answerCallbackQuery"

	var telegramApi = telegramApiBaseUrl + Token + path
	var msg = map[string]string {
		"callback_query_id": queryID,
		"text": text,
	}
	var b bytes.Buffer
	_ = json.NewEncoder(&b).Encode(&msg)
	_, _ = http.Post(telegramApi, "application/json", &b)
}

func BlockEndpoint(webhook string, endpointHash string, authSecret string) bool {
	var payload = map[string] string {
		"endpoint_hash": endpointHash,
		"authSecret": authSecret,
	}
	var b bytes.Buffer
	_ = json.NewEncoder(&b).Encode(&payload)
	response, err := http.Post(webhook+"/block/endpoint", "application/json", &b)
	if err != nil {
		return false
	}
	return response.StatusCode == http.StatusOK
}
