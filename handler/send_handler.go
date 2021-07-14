package handler

import (
	"SecrurumExireBot/bot"
	"SecrurumExireBot/db"
	"SecrurumExireBot/model"
	"encoding/json"
	"net/http"
)

func SendHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct{
		Secret string `json:"secret"`
		Message string `json:"message"`
	}
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
	bot.SendMessage(usr.ChatID, requestBody.Message)
	w.WriteHeader(http.StatusOK)
}
