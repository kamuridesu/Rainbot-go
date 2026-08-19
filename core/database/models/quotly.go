package models

import "time"

type QuotlyFile struct {
	ChatID string `json:"chatId"`
	FileId string `json:"fileId"`
}

type QuotlyMessage struct {
	CreatedAt time.Time `json:"createdAt"`
	StanzaID  string    `json:"stanzaId"`
	ChatID    string    `json:"chatId"`
	FileId    string    `json:"fileId"`
}
