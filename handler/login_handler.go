package handler

import (
	"SecrurumExireBot/bot"
	"SecrurumExireBot/db"
	"SecrurumExireBot/model"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
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
	var usr model.UserModel

	var err = db.DB.Where(&model.UserModel{
		UserID: requestBody.UID,
	}).First(&usr).Error

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if usr.AuthKey != requestBody.Secret {
		bot.SendMessage(usr.ChatID, "Unsuccessful attempt made for login! consider /reKey if it wasn't you.")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	usr.Webhook = requestBody.Webhook
	usr.AuthorizationPayload = requestBody.IdentityString
	if err := db.DB.Updates(&usr).Error; err != nil {
		bot.SendMessage(usr.ChatID, "We were unable to update the webhook and identity string!")
		return
	}


	tokenizer := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": requestBody.UID,
	})

	token, err := tokenizer.SignedString([]byte(bot.Secret))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	bot.SendMessage(usr.ChatID, "Login successful! Webhook: " + requestBody.Webhook)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(token))
}
