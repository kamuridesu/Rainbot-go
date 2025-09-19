package services

import (
	"errors"

	"github.com/kamuridesu/rainbot-go/internal/database/models"
	"github.com/kamuridesu/rainbot-go/internal/database/repositories"
)

type ChatService struct {
	repo repositories.ChatRepository
}

func NewChatService(repo repositories.ChatRepository) *ChatService {
	return &ChatService{repo: repo}
}

func (s *ChatService) GetOrCreateChat(jid string) (*models.Chat, error) {
	chat, err := s.repo.FindById(jid)
	if err != nil {
		return nil, err
	}

	if chat != nil {
		return chat, nil
	}

	if err := s.repo.Create(jid); err != nil {
		return nil, err
	}

	return &models.Chat{
		ChatID:                 jid,
		IsBotEnabled:           1,
		Prefix:                 "/",
		AdminOnly:              0,
		ProfanityFilterEnabled: 0,
		CustomProfanityWords:   "",
		WarnBanThreshold:       4,
	}, nil
}

func (s *ChatService) UpdateChat(chat *models.Chat) error {
	if len(chat.Prefix) != 1 {
		return errors.New("prefixo do chat deve ser de apenas 1 caractere")
	}
	return s.repo.Update(chat)
}
