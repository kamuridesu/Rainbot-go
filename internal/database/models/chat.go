package models

type Chat struct {
	ChatID                 string `json:"chatId"`
	IsBotEnabled           int    `json:"isBotEnabled"`
	Prefix                 string `json:"prefix"`
	AdminOnly              int    `json:"adminOnly"`
	CustomProfanityWords   string `json:"customProfanityWords"`
	ProfanityFilterEnabled int    `json:"profanityFilterEnabled"`
	WarnBanThreshold       int    `json:"warnBanThreshold"`
	AllowAdults            int    `json:"allowAdults"`
	AllowGames             int    `json:"allowGames"`
	AllowFun               int    `json:"allowFun"`
	WelcomeMessage         string `json:"welcomeMessage"`
	CountMessages          int    `json:"countMessages"`
}
