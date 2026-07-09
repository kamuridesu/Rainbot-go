package repositories

import (
	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/database/providers"
)

type QuotlyRepository interface {
	FindAllByChat(chatJid string) ([]*models.QuotlyFile, error)
	FindRandomByChat(chatJid string) (*models.QuotlyFile, error)
	Create(quotly *models.QuotlyFile) error
	Delete(chatJid, fileId string) error
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
