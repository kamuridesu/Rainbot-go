package models

import "time"

type Message struct {
	CreatedAt      time.Time `json:"createdAt"`
	QuotedStanzaID *string   `json:"quotedStanzaId"`
	StanzaID       string    `json:"stanzaId"`
	ChatID         string    `json:"chatId"`
	SenderJID      string    `json:"senderJid"`
	MessageText    string    `json:"messageText"`
}
