package handler

import (
	"SecrurumExireBot/bot"
	"SecrurumExireBot/db"
	"SecrurumExireBot/model"
	"encoding/json"
	"fmt"
	"net/http"
)

func LeakHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct{
		Secret string `json:"secret"`
		Endpoint string `json:"endpoint"`
		SecretName string `json:"secret_name"`
		EndpointHash string `json:"endpoint_hash"`
	}
	fmt.Println(requestBody)
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	uid, err := bot.AuthenticateRequestSecret(requestBody.Secret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var usr model.UserModel
	err = db.DB.Where("user_id = ?", uid).First(&usr).Error
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	msgString := fmt.Sprintf("Attention! we were reported of a leak on endpoint[%s] for the secret[%s]",
		requestBody.Endpoint,
		requestBody.SecretName)

	bot.SendBotCommandMessage(usr.ChatID, msgString,"/block "+requestBody.EndpointHash, "block endpoint")
	w.WriteHeader(http.StatusOK)
}
