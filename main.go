package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os"
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
	dbStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
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
}

func main() {
	port := os.Getenv("PORT")
	http.HandleFunc("/bot", func(w http.ResponseWriter, r *http.Request) {
		var u Update

		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			return
		}

		fmt.Println(u)
		_, _ = w.Write([]byte("Hi"))
	})

	fmt.Println("listening at port... " + port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("error: " + err.Error())
	}
}