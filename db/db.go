package db

import (
	"SecrurumExireBot/bot"
	"SecrurumExireBot/model"
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

var DB *gorm.DB

func InitEnv()  {
	_ = godotenv.Load(".env")
	host := os.Getenv("HOST")
	dbUser := os.Getenv("DB_USER")
	userPass := os.Getenv("DB_USER_PASS")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	bot.Token = os.Getenv("BOT_TOKEN")
	bot.Secret = os.Getenv("SECRET")
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
	err = DB.AutoMigrate(&model.UserModel{})
	if err != nil {
		fmt.Println("Auto migration failed!")
	}
}
