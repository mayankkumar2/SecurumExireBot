package handler

import (
	"SecrurumExireBot/bot"
	"SecrurumExireBot/db"
	"SecrurumExireBot/model"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
)

func BotHandler(_ http.ResponseWriter, r *http.Request) {
	var u model.Update

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		return
	}
	fmt.Println("Received:", u)

	if u.CallbackQuery != nil {
		fmt.Println("Payload:", *u.CallbackQuery)
		command, text := bot.ParseCommand(u.CallbackQuery.Data)
		if command == bot.BlockCommand {
			fmt.Println("Block the endpoint: ",text)
			bot.RespondCallbackQuery(u.CallbackQuery.ID, "Signaling server reported of the action to take")
			var usr model.UserModel
			var err = db.DB.Where("chat_id", u.CallbackQuery.Message.Chat.ID).First(&usr).Error
			if err != nil {
				bot.SendMessage(u.CallbackQuery.Message.Chat.ID, "Oops! something went wrong!")
				return
			}
			success := bot.BlockEndpoint(usr.Webhook, text, usr.AuthorizationPayload)
			if success {
				bot.SendMessage(usr.ChatID, "Hey! signal server reported that the endpoint was blocked!")
			} else {
				bot.SendMessage(usr.ChatID, "Hey! unfortunately some error occurred and we couldn't block the endpoint! Please contact the system admin ASAP!")
			}
		}
	} else {
		command, text := bot.ParseCommand(u.Message.Text)
		if command == bot.StartCommand {
			var c = &model.UserModel{
				ChatID:  u.Message.Chat.ID,
				AuthKey: bot.GenerateUUID() + bot.GenerateUUID(),
				UserID:  uuid.New(),
			}
			err := db.DB.Create(c).Error
			if err != nil {
				bot.SendMessage(c.ChatID, "Hey you might already be a member! Get the required details with /me or try again!")
				return
			}
			s := fmt.Sprintf("Hey! your UID is %s and secret key is %s", c.UserID.String(), c.AuthKey)
			bot.SendMessage(u.Message.Chat.ID, s)
			fmt.Println(text)
		} else if command == bot.MeCommand {
			var usr model.UserModel
			var err = db.DB.Where("chat_id", u.Message.Chat.ID).First(&usr).Error
			if err != nil {
				bot.SendMessage(u.Message.Chat.ID, "Oops! something went wrong!")
				return
			}
			s := fmt.Sprintf("Hey! your UID is %s and secret key is %s", usr.UserID.String(), usr.AuthKey)
			bot.SendMessage(u.Message.Chat.ID, s)
			fmt.Println(text)
		} else if command == bot.ReKeyCommand {
			var usr model.UserModel
			var err = db.DB.Where("chat_id", u.Message.Chat.ID).First(&usr).Error
			if err != nil {
				bot.SendMessage(u.Message.Chat.ID, "Oops! something went wrong!")
				return
			}
			usr.AuthKey = bot.GenerateUUID()+bot.GenerateUUID()
			usr.Webhook = ""
			usr.AuthorizationPayload = ""
			if err := db.DB.Updates(&usr).Error; err != nil {
				bot.SendMessage(u.Message.Chat.ID, "Oops! something went wrong!")
				return
			}
			s := fmt.Sprintf("Hey! your UID is %s and secret key is %s", usr.UserID.String(), usr.AuthKey)
			bot.SendMessage(u.Message.Chat.ID, s)
			fmt.Println(text)
		}
	}
}
