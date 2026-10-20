package repositories

import (
	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/database/providers"
)

type AIModelUsageRepository interface {
	FindAll() ([]*models.AIModelUsage, error)
	Upsert(usage *models.AIModelUsage) error
	Close() error
}

type aiModelUsageRepository struct {
	db *providers.Database
}

func NewAIModelUsageRepository(db *providers.Database) AIModelUsageRepository {
	return &aiModelUsageRepository{db: db}
}

func (r *aiModelUsageRepository) Close() error {
	return r.db.Close()
}

func (r *aiModelUsageRepository) FindAll() ([]*models.AIModelUsage, error) {
	rows, err := r.db.DB.Query(r.db.GetQuery(
		"SELECT name, usedRequestsDay, usedRequestsMin, nextResetDay, nextResetMin FROM ai_model_usage",
	))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usages []*models.AIModelUsage
	for rows.Next() {
		var usage models.AIModelUsage
		if err := rows.Scan(&usage.Name, &usage.UsedRequestsDay, &usage.UsedRequestsMin, &usage.NextResetDay, &usage.NextResetMin); err != nil {
			return nil, err
		}
		usages = append(usages, &usage)
	}
	return usages, rows.Err()
}

func (r *aiModelUsageRepository) Upsert(usage *models.AIModelUsage) error {
	_, err := r.db.DB.Exec(r.db.GetQuery(
		`INSERT INTO ai_model_usage (name, usedRequestsDay, usedRequestsMin, nextResetDay, nextResetMin) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (name) DO UPDATE SET
			usedRequestsDay = excluded.usedRequestsDay,
			usedRequestsMin = excluded.usedRequestsMin,
			nextResetDay = excluded.nextResetDay,
			nextResetMin = excluded.nextResetMin`,
	), usage.Name, usage.UsedRequestsDay, usage.UsedRequestsMin, usage.NextResetDay, usage.NextResetMin)
	return err
}
