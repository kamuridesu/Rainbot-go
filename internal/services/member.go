package services

import (
	"fmt"
	"strings"

	"github.com/kamuridesu/rainbot-go/internal/database/models"
	"github.com/kamuridesu/rainbot-go/internal/database/repositories"
)

type MemberService struct {
	repo repositories.MemberRepository
}

func NewMemberService(repo repositories.MemberRepository) *MemberService {
	return &MemberService{repo: repo}
}

func (s *MemberService) GetOrCreateMember(chatJid, memberJid string) (*models.Member, error) {
	member, err := s.repo.FindByChatAndId(chatJid, memberJid)
	if err != nil {
		return nil, err
	}

	if member != nil {
		return member, nil
	}

	if !strings.HasSuffix(memberJid, "@lid") {
		return nil, fmt.Errorf("invalid user id: %s", memberJid)
	}
	if err := s.repo.Create(chatJid, memberJid); err != nil {
		return nil, err
	}

	return &models.Member{
		ChatID:   chatJid,
		JID:      memberJid,
		Messages: 0,
		Points:   0,
		Warns:    0,
		Silenced: 0,
	}, nil
}

func (s *MemberService) Update(member *models.Member) error {
	return s.repo.Update(member)
}

func (s *MemberService) Close() error {
	return s.repo.Close()
}

func (s *MemberService) GetByChat(chatJid string) ([]*models.Member, error) {
	return s.repo.GetAllByChat(chatJid)
}
