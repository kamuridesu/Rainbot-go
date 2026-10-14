package services

import (
	"errors"
	"time"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/database/repositories"
)

type ChatService struct {
	repo repositories.ChatRepository
}

func NewChatService(repo repositories.ChatRepository) *ChatService {
	return &ChatService{repo: repo}
}

func (s *ChatService) GetOrCreateChat(jid string) (*models.Chat, error) {
	chat, err := s.Get(jid)
	if err != nil {
		return nil, err
	} else if chat != nil {
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
		AllowAdults:            0,
		AllowGames:             1,
		AllowFun:               1,
		WelcomeMessage:         "",
		CountMessages:          1,
		AllowQuote:             1,
		QuoteNMessages:         300,
		AllowOffensiveReplies:  1,
	}, nil
}

func (s *ChatService) Get(chatJid string) (*models.Chat, error) {
	chat, err := s.repo.FindById(chatJid)
	if err != nil {
		return nil, err
	}

	return chat, nil
}

func (s *ChatService) UpdateChat(chat *models.Chat) error {
	if len(chat.Prefix) != 1 {
		return errors.New("prefixo do chat deve ser de apenas 1 caractere")
	}
	return s.repo.Update(chat)
}

func (s *ChatService) MarkLeft(chatJid string) error {
	return s.repo.MarkLeft(chatJid, time.Now().Truncate(time.Second))
}

func (s *ChatService) ClearLeft(chatJid string) error {
	return s.repo.ClearLeft(chatJid)
}

func (s *ChatService) GetLeftBefore(cutoff time.Time) ([]string, error) {
	return s.repo.FindLeftBefore(cutoff)
}

func (s *ChatService) Purge(chatJid string, cutoff time.Time) (bool, error) {
	return s.repo.Purge(chatJid, cutoff)
}

func (s *ChatService) Close() error {
	return s.repo.Close()
}
