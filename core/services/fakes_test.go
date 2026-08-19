package services

import (
	"time"

	"github.com/kamuridesu/rainbot-go/core/database/models"
)

type fakeChatRepo struct {
	findByIdFn func(jid string) (*models.Chat, error)
	updateFn   func(chat *models.Chat) error

	createCalls []string
	createErr   error
	updateCalls []*models.Chat
}

func (f *fakeChatRepo) FindById(jid string) (*models.Chat, error) {
	if f.findByIdFn != nil {
		return f.findByIdFn(jid)
	}
	return nil, nil
}
func (f *fakeChatRepo) Create(jid string) error {
	f.createCalls = append(f.createCalls, jid)
	return f.createErr
}
func (f *fakeChatRepo) Update(chat *models.Chat) error {
	f.updateCalls = append(f.updateCalls, chat)
	if f.updateFn != nil {
		return f.updateFn(chat)
	}
	return nil
}
func (f *fakeChatRepo) Delete(jid string) error { return nil }
func (f *fakeChatRepo) Close() error            { return nil }

type fakeMemberRepo struct {
	createErr         error
	findByChatAndIdFn func(chatJid, memberJid string) (*models.Member, error)
	getAllByChatFn    func(chatJid string) ([]*models.Member, error)
	updateFn          func(member *models.Member) error
	createCalls       [][2]string
}

func (f *fakeMemberRepo) FindByChatAndId(chatJid, memberJid string) (*models.Member, error) {
	if f.findByChatAndIdFn != nil {
		return f.findByChatAndIdFn(chatJid, memberJid)
	}
	return nil, nil
}
func (f *fakeMemberRepo) Create(chatJid, memberJid string) error {
	f.createCalls = append(f.createCalls, [2]string{chatJid, memberJid})
	return f.createErr
}
func (f *fakeMemberRepo) Update(member *models.Member) error {
	if f.updateFn != nil {
		return f.updateFn(member)
	}
	return nil
}
func (f *fakeMemberRepo) GetAllByChat(chatJid string) ([]*models.Member, error) {
	if f.getAllByChatFn != nil {
		return f.getAllByChatFn(chatJid)
	}
	return nil, nil
}
func (f *fakeMemberRepo) Close() error { return nil }

type fakeFilterRepo struct {
	findAllByChatFn func(chatJid string) ([]*models.Filter, error)
	createFn        func(filter *models.Filter) error
	deleteFn        func(chatJid, pattern string) error
}

func (f *fakeFilterRepo) FindAllByChat(chatJid string) ([]*models.Filter, error) {
	if f.findAllByChatFn != nil {
		return f.findAllByChatFn(chatJid)
	}
	return nil, nil
}
func (f *fakeFilterRepo) Create(filter *models.Filter) error {
	if f.createFn != nil {
		return f.createFn(filter)
	}
	return nil
}
func (f *fakeFilterRepo) Delete(chatJid, pattern string) error {
	if f.deleteFn != nil {
		return f.deleteFn(chatJid, pattern)
	}
	return nil
}
func (f *fakeFilterRepo) Close() error { return nil }

type fakeMessageRepo struct {
	createFn          func(msg *models.Message) error
	findByStanzaIDFn  func(stanzaID string) (*models.Message, error)
	findMessagesAfter func(chatId string, since time.Time, limit int) ([]*models.Message, error)
}

func (f *fakeMessageRepo) Create(msg *models.Message) error {
	if f.createFn != nil {
		return f.createFn(msg)
	}
	return nil
}
func (f *fakeMessageRepo) FindByStanzaID(stanzaID string) (*models.Message, error) {
	if f.findByStanzaIDFn != nil {
		return f.findByStanzaIDFn(stanzaID)
	}
	return nil, nil
}
func (f *fakeMessageRepo) FindMessagesAfter(chatId string, since time.Time, limit int) ([]*models.Message, error) {
	if f.findMessagesAfter != nil {
		return f.findMessagesAfter(chatId, since, limit)
	}
	return nil, nil
}
func (f *fakeMessageRepo) StartPartitionManager() {}
func (f *fakeMessageRepo) Close() error           { return nil }

type fakeQuotlyRepo struct {
	findAllByChatFn             func(chatJid string) ([]*models.QuotlyFile, error)
	findRandomByChatFn          func(chatJid string) (*models.QuotlyFile, error)
	createFn                    func(quotly *models.QuotlyFile) error
	deleteFn                    func(chatJid, fileId string) error
	createSentMessageFn         func(msg *models.QuotlyMessage) error
	findSentMessageByStanzaIDFn func(chatJid, stanzaId string) (*models.QuotlyMessage, error)
}

func (f *fakeQuotlyRepo) FindAllByChat(chatJid string) ([]*models.QuotlyFile, error) {
	if f.findAllByChatFn != nil {
		return f.findAllByChatFn(chatJid)
	}
	return nil, nil
}
func (f *fakeQuotlyRepo) FindRandomByChat(chatJid string) (*models.QuotlyFile, error) {
	if f.findRandomByChatFn != nil {
		return f.findRandomByChatFn(chatJid)
	}
	return nil, nil
}
func (f *fakeQuotlyRepo) Create(quotly *models.QuotlyFile) error {
	if f.createFn != nil {
		return f.createFn(quotly)
	}
	return nil
}
func (f *fakeQuotlyRepo) Delete(chatJid, fileId string) error {
	if f.deleteFn != nil {
		return f.deleteFn(chatJid, fileId)
	}
	return nil
}
func (f *fakeQuotlyRepo) CreateSentMessage(msg *models.QuotlyMessage) error {
	if f.createSentMessageFn != nil {
		return f.createSentMessageFn(msg)
	}
	return nil
}
func (f *fakeQuotlyRepo) FindSentMessageByStanzaID(chatJid, stanzaId string) (*models.QuotlyMessage, error) {
	if f.findSentMessageByStanzaIDFn != nil {
		return f.findSentMessageByStanzaIDFn(chatJid, stanzaId)
	}
	return nil, nil
}
func (f *fakeQuotlyRepo) Close() error { return nil }
