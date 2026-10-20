package services

import (
	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/database/repositories"
)

type AIModelUsageService struct {
	repo repositories.AIModelUsageRepository
}

func NewAIModelUsageService(repo repositories.AIModelUsageRepository) *AIModelUsageService {
	return &AIModelUsageService{repo: repo}
}

func (s *AIModelUsageService) GetAll() ([]*models.AIModelUsage, error) {
	return s.repo.FindAll()
}

func (s *AIModelUsageService) Save(usage *models.AIModelUsage) error {
	return s.repo.Upsert(usage)
}

func (s *AIModelUsageService) Close() error {
	return s.repo.Close()
}
