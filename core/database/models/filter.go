package models

type Filter struct {
	ChatID   string `json:"chatId"`
	Pattern  string `json:"pattern"`
	Kind     string `json:"kind"`
	Response string `json:"response"`
}
