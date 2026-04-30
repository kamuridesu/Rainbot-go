package models

import "time"

type Message struct {
	StanzaID       string    `json:"stanzaId"`
	ChatID         string    `json:"chatId"`
	SenderJID      string    `json:"senderJid"`
	MessageText    string    `json:"messageText"`
	QuotedStanzaID *string   `json:"quotedStanzaId"`
	CreatedAt      time.Time `json:"createdAt"`
}
