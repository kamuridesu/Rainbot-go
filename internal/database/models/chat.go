package models

type Chat struct {
	ChatID                 string `json:"chatId"`
	IsBotEnabled           int    `json:"isBotEnabled"`
	Prefix                 string `json:"prefix"`
	AdminOnly              int    `json:"adminOnly"`
	CustomProfanityWords   string `json:"customProfanityWords"`
	ProfanityFilterEnabled int    `json:"profanityFilterEnabled"`
	WarnBanThreshold       int    `json:"warnBanThreshold"`
}
