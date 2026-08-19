package models

type Chat struct {
	CustomProfanityWords   string `json:"customProfanityWords"`
	WelcomeMessage         string `json:"welcomeMessage"`
	Prefix                 string `json:"prefix"`
	ChatID                 string `json:"chatId"`
	WarnBanThreshold       int    `json:"warnBanThreshold"`
	ProfanityFilterEnabled int    `json:"profanityFilterEnabled"`
	AdminOnly              int    `json:"adminOnly"`
	AllowAdults            int    `json:"allowAdults"`
	AllowGames             int    `json:"allowGames"`
	AllowFun               int    `json:"allowFun"`
	IsBotEnabled           int    `json:"isBotEnabled"`
	CountMessages          int    `json:"countMessages"`
	AllowQuote             int    `json:"allowQuote"`
	QuoteNMessages         int    `json:"quoteNMessages"`
	AllowOffensiveReplies  int    `json:"allowOffensiveReplies"`
}
