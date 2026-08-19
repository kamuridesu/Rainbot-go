package quotly

type QuotlyApi struct {
	Url string
}

type QuotlyUserPhoto struct {
	Url string `json:"url"`
}

type QuotlyUser struct {
	FirstName string          `json:"first_name"`
	LastName  string          `json:"last_name"`
	Username  string          `json:"username"`
	Photo     QuotlyUserPhoto `json:"photo"`
	Id        int             `json:"id"`
}

type QuotlyEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

type QuotlyReplyUser struct {
	Name  string          `json:"name"`
	Photo QuotlyUserPhoto `json:"photo"`
	Id    int             `json:"id"`
}

type QuotlyReplyMessage struct {
	From     QuotlyReplyUser `json:"from"`
	Name     string          `json:"name"`
	Text     string          `json:"text"`
	Entities []QuotlyEntity  `json:"entities"`
	ChatId   int             `json:"chatId"`
}

type QuotlyMessage struct {
	ReplyMessage QuotlyReplyMessage `json:"replyMessage"`
	From         QuotlyUser         `json:"from"`
	Text         string             `json:"text"`
	Entities     []QuotlyEntity     `json:"entities"`
	Avatar       bool               `json:"avatar"`
}

type QuotlyRequestBody struct {
	BackgroundColor string          `json:"backgroundColor"`
	EmojiBrand      string          `json:"emojiBrand"`
	Messages        []QuotlyMessage `json:"messages"`
	Width           int             `json:"width"`
	Height          int             `json:"height"`
	Scale           int             `json:"scale"`
}

type QuotlyResponse struct {
	Result struct {
		Image  string `json:"image"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"result"`
	Ok bool `json:"ok"`
}
