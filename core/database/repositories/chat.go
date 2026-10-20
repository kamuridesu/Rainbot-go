package repositories

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	db "github.com/kamuridesu/rainbot-go/core/database/providers"
)

type ChatRepository interface {
	FindById(jid string) (*models.Chat, error)
	Create(jid string) error
	Update(chat *models.Chat) error
	Delete(jid string) error
	MarkLeft(jid string, at time.Time) error
	ClearLeft(jid string) error
	FindLeftBefore(cutoff time.Time) ([]string, error)
	Purge(jid string, cutoff time.Time) (bool, error)
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
		"SELECT chatId, isBotEnabled, prefix, adminOnly, customProfanityWords, profanityFilterEnabled, warnBanThreshold, allowAdults, allowGames, allowFun, welcomeMessage, countMessages, allowQuote, quotlyNMessages, allowOffensiveReplies, allowLLM, botStyle, leftAt FROM chat WHERE chatId = ?",
	), jid)

	var chat models.Chat
	var leftAt sql.NullTime
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
		&chat.AllowQuote,
		&chat.QuoteNMessages,
		&chat.AllowOffensiveReplies,
		&chat.AllowLLM,
		&chat.BotStyle,
		&leftAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if leftAt.Valid {
		chat.LeftAt = &leftAt.Time
	}
	return &chat, nil
}

func (r *chatRepository) Create(jid string) error {
	_, err := r.db.DB.Exec(r.db.GetQuery("INSERT INTO chat (chatId) VALUES (?)"), jid)
	return err
}

func (r *chatRepository) Update(chat *models.Chat) error {
	slog.Warn("updating chat")
	_, err := r.db.DB.Exec(r.db.GetQuery(`
		UPDATE chat SET 
			isBotEnabled = ?, 
			prefix = ?, 
			adminOnly = ?, 
			customProfanityWords = ?, 
			profanityFilterEnabled = ?, 
			warnBanThreshold = ?, 
			allowAdults = ?, 
			allowGames = ?, 
			allowFun = ?, 
			welcomeMessage = ?, 
			countMessages = ?, 
			allowQuote = ?, 
			quotlyNMessages = ?,
			allowOffensiveReplies = ?,
			allowLLM = ?,
			botStyle = ?
		WHERE chatId = ?`),
		chat.IsBotEnabled,
		chat.Prefix,
		chat.AdminOnly,
		chat.CustomProfanityWords,
		chat.ProfanityFilterEnabled,
		chat.WarnBanThreshold,
		chat.AllowAdults,
		chat.AllowGames,
		chat.AllowFun,
		chat.WelcomeMessage,
		chat.CountMessages,
		chat.AllowQuote,
		chat.QuoteNMessages,
		chat.AllowOffensiveReplies,
		chat.AllowLLM,
		chat.BotStyle,
		chat.ChatID,
	)
	return err
}

func (r *chatRepository) Delete(jid string) error {
	_, err := r.db.DB.Exec(r.db.GetQuery("DELETE FROM chat WHERE chatId = ?"), jid)
	return err
}

func (r *chatRepository) MarkLeft(jid string, at time.Time) error {
	_, err := r.db.DB.Exec(r.db.GetQuery("UPDATE chat SET leftAt = ? WHERE chatId = ? AND leftAt IS NULL"), at.UTC(), jid)
	return err
}

func (r *chatRepository) ClearLeft(jid string) error {
	_, err := r.db.DB.Exec(r.db.GetQuery("UPDATE chat SET leftAt = NULL WHERE chatId = ? AND leftAt IS NOT NULL"), jid)
	return err
}

func (r *chatRepository) FindLeftBefore(cutoff time.Time) ([]string, error) {
	rows, err := r.db.DB.Query(r.db.GetQuery("SELECT chatId FROM chat WHERE leftAt IS NOT NULL AND leftAt <= ?"), cutoff.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *chatRepository) Purge(jid string, cutoff time.Time) (bool, error) {
	tx, err := r.db.DB.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var found int
	err = tx.QueryRow(r.db.GetQuery(
		"SELECT 1 FROM chat WHERE chatId = ? AND leftAt IS NOT NULL AND leftAt <= ?",
	), jid, cutoff.UTC()).Scan(&found)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	tables := []string{"quotly_message", "quotly", "filter", "member", "messages"}
	if r.db.Driver == "postgres" {
		tables = append(tables, "token")
	}
	tables = append(tables, "chat")

	for _, table := range tables {
		if _, err := tx.Exec(r.db.GetQuery("DELETE FROM "+table+" WHERE chatId = ?"), jid); err != nil {
			return false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
