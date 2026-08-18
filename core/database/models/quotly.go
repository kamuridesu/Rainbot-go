package models

import "time"

type QuotlyFile struct {
	ChatID string `json:"chatId"`
	FileId string `json:"fileId"`
}

type QuotlyMessage struct {
	StanzaID  string    `json:"stanzaId"`
	ChatID    string    `json:"chatId"`
	FileId    string    `json:"fileId"`
	CreatedAt time.Time `json:"createdAt"`
}
