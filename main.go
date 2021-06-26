package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"net/url"
	"os"
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

var DB *gorm.DB
var token string
var secret string
type UserModel struct {
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey"`
	ChatID int `json:"chat_id" gorm:"unique"`
	AuthKey string `json:"-"`
	Webhook string `json:"web_hook"`
	AuthorizationPayload string `json:"authorization_payload"`
}

func initEnv()  {
	_ = godotenv.Load(".env")
	host := os.Getenv("HOST")
	dbUser := os.Getenv("DB_USER")
	userPass := os.Getenv("DB_USER_PASS")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	token = os.Getenv("BOT_TOKEN")
	secret = os.Getenv("SECRET")
	dbStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
		host,
		dbUser,
		userPass,
		dbName,
		port)
	var err error
	DB, err = gorm.Open(postgres.Open(dbStr), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = DB.AutoMigrate(&UserModel{})
	if err != nil {
		fmt.Println("Auto migration failed!")
	}

}

type Update struct {
	UpdateID int `json:"update_id"`
	Message Message `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type CallbackQuery struct {
	ID string `json:"id"`
	Chat Chat `json:"chat"`
	ChatInstance string `json:"chat_instance"`
	Data string `json:"data"`
}

type Message struct {
	Text string `json:"text"`
	Chat Chat `json:"chat"`
}
type Chat struct {
	ID int `json:"id"`
}

func main() {
	initEnv()

	port := os.Getenv("PORT")
	http.HandleFunc("/bot", func(w http.ResponseWriter, r *http.Request) {
		var u Update

		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			return
		}
		fmt.Println("Received:", u)
		if u.CallbackQuery != nil {
			command, text := ParseCommand(u.CallbackQuery.Data)
			if command == BlockCommand {
				fmt.Println("Block the endpoint: ",text)
				RespondCallbackQuery(u.CallbackQuery.ID, "Signaling server reported of the action to take")
			}
		} else {
			command, text := ParseCommand(u.Message.Text)
			if command == StartCommand {
				var c = &UserModel{
					ChatID:  u.Message.Chat.ID,
					AuthKey: GenerateUUID(),
					UserID:  uuid.New(),
				}
				err := DB.Create(c).Error
				if err != nil {
					SendMessage(c.ChatID, "Oops! something went wrong!")
					return
				}
				s := fmt.Sprintf("Hey! your UID is %s and secret key is %s", c.UserID.String(), c.AuthKey)
				SendMessage(u.Message.Chat.ID, s)
				fmt.Println(text)
			} else if command == MeCommand {
				var usr UserModel
				var err = DB.Where("chat_id", u.Message.Chat.ID).First(&usr).Error
				if err != nil {
					SendMessage(u.Message.Chat.ID, "Oops! something went wrong!")
					return
				}
				s := fmt.Sprintf("Hey! your UID is %s and secret key is %s", usr.UserID.String(), usr.AuthKey)
				SendMessage(u.Message.Chat.ID, s)
				fmt.Println(text)
			} else if command == ReKeyCommand {
				var usr UserModel
				var err = DB.Where("chat_id", u.Message.Chat.ID).First(&usr).Error
				if err != nil {
					SendMessage(u.Message.Chat.ID, "Oops! something went wrong!")
					return
				}
				usr.AuthKey = GenerateUUID()
				usr.Webhook = ""
				usr.AuthorizationPayload = ""
				if err := DB.Updates(&usr).Error; err != nil {
					SendMessage(u.Message.Chat.ID, "Oops! something went wrong!")
					return
				}
				s := fmt.Sprintf("Hey! your UID is %s and secret key is %s", usr.UserID.String(), usr.AuthKey)
				SendMessage(u.Message.Chat.ID, s)
				fmt.Println(text)
			}
		}
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var requestBody struct{
			UID uuid.UUID `json:"uid"`
			Secret string `json:"secret"`
			Webhook string `json:"webhook"`
			IdentityString string `json:"identity_string"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Println(requestBody)
		var usr UserModel

		var err = DB.Where(&UserModel{
			UserID: requestBody.UID,
		}).First(&usr).Error

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if usr.AuthKey != requestBody.Secret {
			SendMessage(usr.ChatID, "Unsuccessful attempt made for login! consider /reKey if it wasn't you.")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		usr.Webhook = requestBody.Webhook
		usr.AuthorizationPayload = requestBody.IdentityString
		if err := DB.Updates(&usr).Error; err != nil {
			SendMessage(usr.ChatID, "We were unable to update the webhook and identity string!")
			return
		}


		tokenizer := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"uid": requestBody.UID,
		})

		token, err := tokenizer.SignedString([]byte(secret))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(err)
			return
		}

		SendMessage(usr.ChatID, "Login successful! Webhook: " + requestBody.Webhook)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(token))
	})
	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		var requestBody struct{
			Secret string `json:"secret"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		uid, err := AuthenticateRequestSecret(requestBody.Secret)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var usr UserModel
		err = DB.Where("user_id = ?", uid).First(&usr).Error
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		SendMessage(usr.ChatID, requestBody.Message)
		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("listening at port... " + port)

	err := http.ListenAndServe("0.0.0.0:"+port, nil)
	if err != nil {
		fmt.Println("error: " + err.Error())
	}
}


var startCommand = "/start"
var startCommandLen = len(startCommand)

var meCommand = "/me"
var meCommandLen = len(meCommand)

var reKeyCommand = "/reKey"
var reKeyCommandLen = len(reKeyCommand)

var blockCommand = "/block"
var blockCommandLen = len(blockCommand)

func ParseCommand(text string) (int, string) {
	startCommandRegex, _ := regexp.Compile("^/start")
	meCommandRegex, _ := regexp.Compile("^/me")
	reKeyCommandRegex, _ := regexp.Compile("^/reKey")
	blockCommandRegex, _ := regexp.Compile("^/block")
	if startCommandRegex.MatchString(text) {
		return StartCommand, text[startCommandLen:]
	} else if meCommandRegex.MatchString(text) {
		return MeCommand, text[meCommandLen:]
	} else if reKeyCommandRegex.MatchString(text) {
		return ReKeyCommand, text[reKeyCommandLen:]
	} else if blockCommandRegex.MatchString(text) {
		return BlockCommand, text[blockCommandLen+1:]
	}
	return -1, ""
}

func SendMessage(chatID int, text string) {
	const telegramApiBaseUrl string = "https://api.telegram.org/bot"
	const telegramApiSendMessage string = "/sendMessage"

	var telegramApi = telegramApiBaseUrl + token + telegramApiSendMessage
	response, err := http.PostForm(telegramApi, url.Values{
		"chat_id": {strconv.Itoa(chatID)},
		"text": {text},
	})
	if err != nil {
		return
	}
	defer response.Body.Close()
}

func SendBotCommandMessage(chatID int, text string, command string, buttonText string) {
	const telegramApiBaseUrl string = "https://api.telegram.org/bot"
	const telegramApiSendMessage string = "/sendMessage"

	var telegramApi = telegramApiBaseUrl + token + telegramApiSendMessage
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
		return
	}
	defer response.Body.Close()
}




func GenerateUUID() string {
	uid := uuid.New().String()
	return strings.ReplaceAll(uid, "-", "")
}

func AuthenticateRequestSecret(sec string) (string, error){
	token, err := jwt.Parse(sec, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println(claims["uid"].(string))
		return claims["uid"].(string), nil
	} else {
		return "", errors.New("error: token invalid")
	}
}

func RespondCallbackQuery(queryID, text string) {
	const telegramApiBaseUrl string = "https://api.telegram.org/bot"
	const path string = "/answerCallbackQuery"

	var telegramApi = telegramApiBaseUrl + token + path
	var msg = map[string]string {
		"callback_query_id": queryID,
		"text": text,
	}
	var b bytes.Buffer
	_ = json.NewEncoder(&b).Encode(&msg)
	_, _ = http.Post(telegramApi, "application/json", &b)
}