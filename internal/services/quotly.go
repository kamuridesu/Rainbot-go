package services

import (
	"github.com/kamuridesu/rainbot-go/internal/database/models"
	"github.com/kamuridesu/rainbot-go/internal/database/repositories"
)

type QuotlyService struct {
	repo repositories.QuotlyRepository
}

func NewQuotlyService(repo repositories.QuotlyRepository) *QuotlyService {
	return &QuotlyService{repo: repo}
}

func (s *QuotlyService) SaveQuotly(quotly *models.QuotlyFile) error {
	return s.repo.Create(quotly)
}

func (s *QuotlyService) GetAllByChat(chatJid string) ([]*models.QuotlyFile, error) {
	return s.repo.FindAllByChat(chatJid)
}

func (s *QuotlyService) GetRandomByChat(chatJid string) (*models.QuotlyFile, error) {
	return s.repo.FindRandomByChat(chatJid)
}

func (s *QuotlyService) DeleteQuotly(chatJid, fileId string) error {
	return s.repo.Delete(chatJid, fileId)
}

func (s *QuotlyService) Close() error {
	return s.repo.Close()
}
