package services

import (
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
