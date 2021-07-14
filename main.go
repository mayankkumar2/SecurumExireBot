package main

import (
	"SecrurumExireBot/db"
	"SecrurumExireBot/handler"
	"fmt"
	"net/http"
	"os"
)

func main() {
	db.InitEnv()
	port := os.Getenv("PORT")
	http.HandleFunc("/report/leak", handler.LeakHandler)
	http.HandleFunc("/bot", handler.BotHandler)
	http.HandleFunc("/login", handler.LoginHandler)
	http.HandleFunc("/send", handler.SendHandler)

	fmt.Println("listening at port... " + port)

	err := http.ListenAndServe("0.0.0.0:"+port, nil)
	if err != nil {
		fmt.Println("error: " + err.Error())
	}
}

