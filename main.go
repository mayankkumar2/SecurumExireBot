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
	"strconv"
	"strings"
)

const (
	StartCommand int = 1
)

var DB *gorm.DB
var token string
type UserModel struct {
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid"`
	ChatID int `json:"chat_id" gorm:"unique"`
	AuthKey string `json:"-"`
	WebHook string `json:"web_hook"`
}

func init()  {
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
	DB, err := gorm.Open(postgres.Open(dbStr), &gorm.Config{})
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
	port := os.Getenv("PORT")
	http.HandleFunc("/bot", func(w http.ResponseWriter, r *http.Request) {
		var u Update

		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			return
		}

		command, text := ParseCommand(u.Message.Text)
		if command == StartCommand {
			var c = &UserModel{
				ChatID: u.Message.Chat.ID,
				AuthKey: GenerateUUID(),
				UserID: uuid.New(),
			}

			s := fmt.Sprintf("Hey! your UID is %s and secret key is %s", c.UserID.String(), c.AuthKey)
			SendMessage(u.Message.Chat.ID, s)
			fmt.Println(text)
		}


		fmt.Println(u)
	})

	fmt.Println("listening at port... " + port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("error: " + err.Error())
	}
}


var startCommand = "/start"
var startCommandLen = len("/start")

func ParseCommand(text string) (int, string) {
	if text[:startCommandLen] == startCommand {
		return StartCommand, text[startCommandLen:]
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