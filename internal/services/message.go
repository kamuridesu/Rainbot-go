package services

import (
	"time"

	"github.com/kamuridesu/rainbot-go/internal/database/models"
	"github.com/kamuridesu/rainbot-go/internal/database/repositories"
)

type MessageService struct {
	repo repositories.MessageRepository
}

func NewMessageService(repo repositories.MessageRepository) *MessageService {
	return &MessageService{repo: repo}
}

func (s *MessageService) SaveMessage(msg *models.Message) error {
	return s.repo.Create(msg)
}

func (s *MessageService) GetMessage(stanzaID string) (*models.Message, error) {
	return s.repo.FindByStanzaID(stanzaID)
}

func (s *MessageService) GetMessageRange(chatId string, since time.Time, limit int) ([]*models.Message, error) {
	return s.repo.FindMessagesAfter(chatId, since, limit)
}

func (s *MessageService) Close() error {
	return s.repo.Close()
}
