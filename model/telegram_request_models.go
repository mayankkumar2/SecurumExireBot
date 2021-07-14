package model

type Update struct {
	UpdateID int `json:"update_id"`
	Message Message `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type CallbackQuery struct {
	ID string `json:"id"`
	Message Message `json:"message"`
	ChatInstance string `json:"chat_instance"`
	Data string `json:"data"`
}

type Message struct {
	Text string `json:"text"`
	Chat Chat `json:"chat"`
}
type Chat struct {
	ID int `json:"id"`
}
