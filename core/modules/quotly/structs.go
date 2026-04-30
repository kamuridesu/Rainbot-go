package quotly

type QuotlyApi struct {
	Url string
}

type QuotlyUserPhoto struct {
	Url string `json:"url"`
}

type QuotlyUser struct {
	Id        int             `json:"id"`
	FirstName string          `json:"first_name"`
	LastName  string          `json:"last_name"`
	Username  string          `json:"username"`
	Photo     QuotlyUserPhoto `json:"photo"`
}

type QuotlyEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

type QuotlyReplyUser struct {
	Id    int             `json:"id"`
	Name  string          `json:"name"`
	Photo QuotlyUserPhoto `json:"photo"`
}

type QuotlyReplyMessage struct {
	Name     string          `json:"name"`
	Text     string          `json:"text"`
	Entities []QuotlyEntity  `json:"entities"`
	ChatId   int             `json:"chatId"`
	From     QuotlyReplyUser `json:"from"`
}

type QuotlyMessage struct {
	From         QuotlyUser         `json:"from"`
	Text         string             `json:"text"`
	Entities     []QuotlyEntity     `json:"entities"`
	Avatar       bool               `json:"avatar"`
	ReplyMessage QuotlyReplyMessage `json:"replyMessage"`
}

type QuotlyRequestBody struct {
	BackgroundColor string          `json:"backgroundColor"`
	Width           int             `json:"width"`
	Height          int             `json:"height"`
	Scale           int             `json:"scale"`
	EmojiBrand      string          `json:"emojiBrand"`
	Messages        []QuotlyMessage `json:"messages"`
}

type QuotlyResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		Image  string `json:"image"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"result"`
}
