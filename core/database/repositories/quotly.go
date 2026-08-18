package repositories

import (
	"database/sql"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/database/providers"
)

const maxSentMessagesPerFile = 5

type QuotlyRepository interface {
	FindAllByChat(chatJid string) ([]*models.QuotlyFile, error)
	FindRandomByChat(chatJid string) (*models.QuotlyFile, error)
	Create(quotly *models.QuotlyFile) error
	Delete(chatJid, fileId string) error
	CreateSentMessage(msg *models.QuotlyMessage) error
	FindSentMessageByStanzaID(chatJid, stanzaId string) (*models.QuotlyMessage, error)
	Close() error
}

type quotlyRepository struct {
	db *providers.Database
}

func NewQuotlyRepository(db *providers.Database) QuotlyRepository {
	return &quotlyRepository{db: db}
}

func (r *quotlyRepository) Close() error {
	return r.db.Close()
}

func (r *quotlyRepository) FindAllByChat(chatJid string) ([]*models.QuotlyFile, error) {
	rows, err := r.db.DB.Query(r.db.GetQuery(
		"SELECT chatId, fileId FROM quotly WHERE chatId = ?",
	), chatJid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotlies []*models.QuotlyFile
	for rows.Next() {
		var quotly models.QuotlyFile
		if err := rows.Scan(&quotly.ChatID, &quotly.FileId); err != nil {
			return nil, err
		}
		quotlies = append(quotlies, &quotly)
	}
	return quotlies, nil
}

func (r *quotlyRepository) FindRandomByChat(chatJid string) (*models.QuotlyFile, error) {
	var quotly models.QuotlyFile

	err := r.db.DB.QueryRow(r.db.GetQuery(
		"SELECT chatId, fileId FROM quotly WHERE chatId = ? ORDER BY RANDOM() LIMIT 1",
	), chatJid).Scan(&quotly.ChatID, &quotly.FileId)

	if err != nil {
		return nil, err
	}

	return &quotly, nil
}

func (r *quotlyRepository) Create(quotly *models.QuotlyFile) error {
	_, err := r.db.DB.Exec(r.db.GetQuery(
		"INSERT INTO quotly (chatId, fileId) VALUES (?, ?)",
	), quotly.ChatID, quotly.FileId)
	return err
}

func (r *quotlyRepository) Delete(chatJid, fileId string) error {
	_, err := r.db.DB.Exec(r.db.GetQuery(
		"DELETE FROM quotly WHERE chatId = ? AND fileId = ?",
	), chatJid, fileId)
	return err
}

func (r *quotlyRepository) CreateSentMessage(msg *models.QuotlyMessage) error {
	tx, err := r.db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(r.db.GetQuery(
		"INSERT INTO quotly_message (stanzaId, chatId, fileId, createdAt) VALUES (?, ?, ?, ?)",
	), msg.StanzaID, msg.ChatID, msg.FileId, msg.CreatedAt)
	if err != nil {
		return err
	}

	_, err = tx.Exec(r.db.GetQuery(
		`DELETE FROM quotly_message WHERE chatId = ? AND fileId = ? AND stanzaId NOT IN (
			SELECT stanzaId FROM quotly_message WHERE chatId = ? AND fileId = ? ORDER BY createdAt DESC LIMIT ?
		)`,
	), msg.ChatID, msg.FileId, msg.ChatID, msg.FileId, maxSentMessagesPerFile)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *quotlyRepository) FindSentMessageByStanzaID(chatJid, stanzaId string) (*models.QuotlyMessage, error) {
	var msg models.QuotlyMessage

	err := r.db.DB.QueryRow(r.db.GetQuery(
		"SELECT stanzaId, chatId, fileId, createdAt FROM quotly_message WHERE chatId = ? AND stanzaId = ?",
	), chatJid, stanzaId).Scan(&msg.StanzaID, &msg.ChatID, &msg.FileId, &msg.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &msg, nil
}
