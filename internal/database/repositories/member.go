package repositories

import (
	"database/sql"

	"github.com/kamuridesu/rainbot-go/internal/database/models"
	"github.com/kamuridesu/rainbot-go/internal/database/providers"
)

type MemberRepository interface {
	FindByChatAndId(chatJid, memberJid string) (*models.Member, error)
	Create(chatJid, memberJid string) error
	Update(member *models.Member) error
	Close() error
}

type memberRepository struct {
	db *providers.Database
}

func NewMemberRepository(db *providers.Database) MemberRepository {
	return &memberRepository{db: db}
}

func (r *memberRepository) Close() error {
	return r.db.Close()
}

func (r *memberRepository) FindByChatAndId(chatJid, memberJid string) (*models.Member, error) {
	row := r.db.DB.QueryRow(r.db.GetQuery(
		"SELECT chatId, jid, messages, points, warns, silenced FROM member WHERE chatId = ? AND jid = ?",
	), chatJid, memberJid)

	var member models.Member
	err := row.Scan(&member.ChatID, &member.JID, &member.Messages, &member.Points, &member.Warns, &member.Silenced)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

func (r *memberRepository) Create(chatJid, memberJid string) error {
	_, err := r.db.DB.Exec(r.db.GetQuery("INSERT INTO member (chatId, jid) VALUES (?, ?)"), chatJid, memberJid)
	return err
}

func (r *memberRepository) Update(member *models.Member) error {
	_, err := r.db.DB.Exec(r.db.GetQuery(
		"UPDATE member SET warns = ?, points = ?, messages = ?, silenced = ? WHERE chatId = ? AND jid = ?",
	), member.Warns, member.Points, member.Messages, member.Silenced, member.ChatID, member.JID)
	return err
}
