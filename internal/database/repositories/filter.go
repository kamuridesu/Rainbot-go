package repositories

import (
	"github.com/kamuridesu/rainbot-go/internal/database/models"
	"github.com/kamuridesu/rainbot-go/internal/database/providers"
)

type FilterRepository interface {
	FindAllByChat(chatJid string) ([]*models.Filter, error)
	Create(filter *models.Filter) error
	Delete(chatjid, pattern string) error
}

type filterRepository struct {
	db *providers.Database
}

func NewFilterRepository(db *providers.Database) FilterRepository {
	return &filterRepository{db: db}
}

func (r *filterRepository) FindAllByChat(chatJid string) ([]*models.Filter, error) {
	rows, err := r.db.DB.Query(r.db.GetQuery(
		"SELECT chatId, pattern, kind, response FROM filter WHERE chatId = ?",
	), chatJid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var filters []*models.Filter
	for rows.Next() {
		var filter models.Filter
		if err := rows.Scan(&filter.ChatID, &filter.Pattern, &filter.Kind, &filter.Response); err != nil {
			return nil, err
		}
		filters = append(filters, &filter)
	}
	return filters, nil
}

func (r *filterRepository) Create(filter *models.Filter) error {
	_, err := r.db.DB.Exec(r.db.GetQuery(
		"INSERT INTO filter (chatId, pattern, kind, response) VALUES (?, ?, ?, ?)",
	), filter.ChatID, filter.Pattern, filter.Kind, filter.Response)
	return err
}

func (r *filterRepository) Delete(chatJid, pattern string) error {
	_, err := r.db.DB.Exec(r.db.GetQuery("DELETE FROM filter WHERE chatId = ? AND pattern = ?"), chatJid, pattern)
	return err
}
