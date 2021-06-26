package main

import (
	"encoding/json"
	"fmt"
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

)

var DB *gorm.DB
var token string
type UserModel struct {
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid"`
	ChatID int `json:"chat_id" gorm:"unique"`
	AuthKey string `json:"-"`
	WebHook string `json:"web_hook"`
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

		command, text := ParseCommand(u.Message.Text)
		if command == StartCommand {
			var c = &UserModel{
				ChatID: u.Message.Chat.ID,
				AuthKey: GenerateUUID(),
				UserID: uuid.New(),
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
			var err = DB.First(&usr).Where("chat_id", u.Message.Chat.ID).Error
			if err != nil {
				SendMessage(u.Message.Chat.ID, "Oops! something went wrong!")
				return
			}
			s := fmt.Sprintf("Hey! your UID is %s and secret key is %s", usr.UserID.String(), usr.AuthKey)
			SendMessage(u.Message.Chat.ID, s)
			fmt.Println(text)
		}


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
func ParseCommand(text string) (int, string) {
	startCommandRegex, _ := regexp.Compile("^/start")
	meCommandRegex, _ := regexp.Compile("^/start")

	if startCommandRegex.MatchString(text) {
		return StartCommand, text[startCommandLen:]
	} else if meCommandRegex.MatchString(text) {
		return MeCommand, text[meCommandLen:]
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

func GenerateUUID() string {
	uid := uuid.New().String()
	return strings.ReplaceAll(uid, "-", "")
}