package services

import (
	"github.com/kamuridesu/rainbot-go/internal/database/models"
	"github.com/kamuridesu/rainbot-go/internal/database/repositories"
)

type FilterService struct {
	repo repositories.FilterRepository
}

func NewFilterRepository(repo repositories.FilterRepository) *FilterService {
	return &FilterService{repo: repo}
}

func (s *FilterService) GetFilters(chatJid string) ([]*models.Filter, error) {
	return s.repo.FindAllByChat(chatJid)
}

func (s *FilterService) NewFilter(filter *models.Filter) error {
	return s.repo.Create(filter)
}

func (s *FilterService) Delete(chatJid, pattern string) error {
	return s.repo.Delete(chatJid, pattern)
}
