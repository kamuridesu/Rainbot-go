package models

import (
	"slices"
	"strings"
	"time"
)

type Chat struct {
	ChatID                 string     `json:"chatId"`
	IsBotEnabled           int        `json:"isBotEnabled"`
	Prefix                 string     `json:"prefix"`
	AdminOnly              int        `json:"adminOnly"`
	CustomProfanityWords   string     `json:"customProfanityWords"`
	ProfanityFilterEnabled int        `json:"profanityFilterEnabled"`
	WarnBanThreshold       int        `json:"warnBanThreshold"`
	DisabledCategories     string     `json:"disabledCategories"`
	WelcomeMessage         string     `json:"welcomeMessage"`
	CountMessages          int        `json:"countMessages"`
	AllowQuote             int        `json:"allowQuote"`
	QuoteNMessages         int        `json:"quoteNMessages"`
	AllowOffensiveReplies  int        `json:"allowOffensiveReplies"`
	AllowLLM               int        `json:"allowLLM"`
	BotStyle               string     `json:"botStyle"`
	LeftAt                 *time.Time `json:"leftAt"`
}

func (c *Chat) GetDisabledCategories() []string {
	var categories []string
	for category := range strings.SplitSeq(c.DisabledCategories, ",") {
		if category = strings.TrimSpace(category); category != "" {
			categories = append(categories, category)
		}
	}
	return categories
}

func (c *Chat) IsCategoryDisabled(name string) bool {
	return slices.ContainsFunc(c.GetDisabledCategories(), func(category string) bool {
		return strings.EqualFold(category, name)
	})
}
