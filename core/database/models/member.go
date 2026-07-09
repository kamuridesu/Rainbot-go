package models

type Member struct {
	ChatID   string `json:"chatId"`
	JID      string `json:"jid"`
	Messages int    `json:"messages"`
	Points   int    `json:"points"`
	Warns    int    `json:"warns"`
	Silenced int    `json:"silenced"`
}
