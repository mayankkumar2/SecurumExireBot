package model

import "github.com/google/uuid"

type UserModel struct {
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey"`
	ChatID int `json:"chat_id" gorm:"unique"`
	AuthKey string `json:"-"`
	Webhook string `json:"web_hook"`
	AuthorizationPayload string `json:"authorization_payload"`
}
