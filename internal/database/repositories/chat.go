package repositories

import (
	"database/sql"

	"github.com/kamuridesu/rainbot-go/internal/database/models"
	db "github.com/kamuridesu/rainbot-go/internal/database/providers"
)

type ChatRepository interface {
	FindById(jid string) (*models.Chat, error)
	Create(jid string) error
	Update(chat *models.Chat) error
	Delete(jid string) error
	Close() error
}

type chatRepository struct {
	db *db.Database
}

func NewChatRepository(db *db.Database) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) Close() error {
	return r.db.Close()
}

func (r *chatRepository) FindById(jid string) (*models.Chat, error) {
	row := r.db.DB.QueryRow(r.db.GetQuery(
		"SELECT chatId, isBotEnabled, prefix, adminOnly, customProfanityWords, profanityFilterEnabled, warnBanThreshold, allowAdults, allowGames, allowFun, welcomeMessage, countMessages FROM chat WHERE chatId = ?",
	), jid)

	var chat models.Chat
	err := row.Scan(
		&chat.ChatID,
		&chat.IsBotEnabled,
		&chat.Prefix,
		&chat.AdminOnly,
		&chat.CustomProfanityWords,
		&chat.ProfanityFilterEnabled,
		&chat.WarnBanThreshold,
		&chat.AllowAdults,
		&chat.AllowGames,
		&chat.AllowFun,
		&chat.WelcomeMessage,
		&chat.CountMessages,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
	}
	return &chat, nil
}

func (r *chatRepository) Create(jid string) error {
	_, err := r.db.DB.Exec(r.db.GetQuery("INSERT INTO chat (chatId) VALUES (?)"), jid)
	return err
}

func (r *chatRepository) Update(chat *models.Chat) error {
	_, err := r.db.DB.Exec(r.db.GetQuery("UPDATE chat SET isBotEnabled = ?, prefix = ?, adminOnly = ?, customProfanityWords = ?, profanityFilterEnabled = ?, warnBanThreshold = ?, allowAdults = ?, allowGames = ?, allowFun = ?, welcomeMessage = ?, countMessages = ? WHERE chatId = ?"), chat.IsBotEnabled, chat.Prefix, chat.AdminOnly, chat.CustomProfanityWords, chat.ProfanityFilterEnabled, chat.WarnBanThreshold, chat.AllowAdults, chat.AllowGames, chat.AllowFun, chat.WelcomeMessage, chat.CountMessages, chat.ChatID)
	return err
}

func (r *chatRepository) Delete(jid string) error {
	_, err := r.db.DB.Exec(r.db.GetQuery("DELETE FROM chat WHERE chatId = ?"), jid)
	return err
}
