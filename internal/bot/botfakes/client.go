package botfakes

import (
	"context"

	"github.com/kamuridesu/rainbot-go/internal/bot"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
)

type SentMessage struct {
	Chat    types.JID
	Message *waE2E.Message
}

type FakeClient struct {
	SendMessageFunc             func(ctx context.Context, to types.JID, message *waE2E.Message, extra ...whatsmeow.SendRequestExtra) (whatsmeow.SendResponse, error)
	BuildRevokeFunc             func(chat, sender types.JID, id types.MessageID) *waE2E.Message
	BuildReactionFunc           func(chat, sender types.JID, id types.MessageID, reaction string) *waE2E.Message
	BuildEditFunc               func(chat types.JID, id types.MessageID, newContent *waE2E.Message) *waE2E.Message
	UploadFunc                  func(ctx context.Context, plaintext []byte, appInfo whatsmeow.MediaType) (whatsmeow.UploadResponse, error)
	DownloadFunc                func(ctx context.Context, msg whatsmeow.DownloadableMessage) ([]byte, error)
	GetGroupInfoFunc            func(ctx context.Context, jid types.JID) (*types.GroupInfo, error)
	GetProfilePictureInfoFunc   func(ctx context.Context, jid types.JID, params *whatsmeow.GetProfilePictureParams) (*types.ProfilePictureInfo, error)
	GetJoinedGroupsFunc         func(ctx context.Context) ([]*types.GroupInfo, error)
	UpdateGroupParticipantsFunc func(ctx context.Context, jid types.JID, participantChanges []types.JID, action whatsmeow.ParticipantChange) ([]types.GroupParticipant, error)
	StoreFunc                   func() *store.Device
	IsConnectedFunc             func() bool

	SentMessages             []SentMessage
	UpdatedGroupParticipants []types.JID
	DisconnectCalled         bool
}

func (f *FakeClient) SendMessage(ctx context.Context, to types.JID, message *waE2E.Message, extra ...whatsmeow.SendRequestExtra) (whatsmeow.SendResponse, error) {
	f.SentMessages = append(f.SentMessages, SentMessage{Chat: to, Message: message})
	if f.SendMessageFunc != nil {
		return f.SendMessageFunc(ctx, to, message, extra...)
	}
	return whatsmeow.SendResponse{}, nil
}

func (f *FakeClient) BuildRevoke(chat, sender types.JID, id types.MessageID) *waE2E.Message {
	if f.BuildRevokeFunc != nil {
		return f.BuildRevokeFunc(chat, sender, id)
	}
	return &waE2E.Message{}
}

func (f *FakeClient) BuildReaction(chat, sender types.JID, id types.MessageID, reaction string) *waE2E.Message {
	if f.BuildReactionFunc != nil {
		return f.BuildReactionFunc(chat, sender, id, reaction)
	}
	return &waE2E.Message{}
}

func (f *FakeClient) BuildEdit(chat types.JID, id types.MessageID, newContent *waE2E.Message) *waE2E.Message {
	if f.BuildEditFunc != nil {
		return f.BuildEditFunc(chat, id, newContent)
	}
	return &waE2E.Message{}
}

func (f *FakeClient) Upload(ctx context.Context, plaintext []byte, appInfo whatsmeow.MediaType) (whatsmeow.UploadResponse, error) {
	if f.UploadFunc != nil {
		return f.UploadFunc(ctx, plaintext, appInfo)
	}
	return whatsmeow.UploadResponse{}, nil
}

func (f *FakeClient) Download(ctx context.Context, msg whatsmeow.DownloadableMessage) ([]byte, error) {
	if f.DownloadFunc != nil {
		return f.DownloadFunc(ctx, msg)
	}
	return nil, nil
}

func (f *FakeClient) GetGroupInfo(ctx context.Context, jid types.JID) (*types.GroupInfo, error) {
	if f.GetGroupInfoFunc != nil {
		return f.GetGroupInfoFunc(ctx, jid)
	}
	return &types.GroupInfo{}, nil
}

func (f *FakeClient) GetProfilePictureInfo(ctx context.Context, jid types.JID, params *whatsmeow.GetProfilePictureParams) (*types.ProfilePictureInfo, error) {
	if f.GetProfilePictureInfoFunc != nil {
		return f.GetProfilePictureInfoFunc(ctx, jid, params)
	}
	return nil, nil
}

func (f *FakeClient) GetJoinedGroups(ctx context.Context) ([]*types.GroupInfo, error) {
	if f.GetJoinedGroupsFunc != nil {
		return f.GetJoinedGroupsFunc(ctx)
	}
	return nil, nil
}

func (f *FakeClient) UpdateGroupParticipants(ctx context.Context, jid types.JID, participantChanges []types.JID, action whatsmeow.ParticipantChange) ([]types.GroupParticipant, error) {
	f.UpdatedGroupParticipants = append(f.UpdatedGroupParticipants, participantChanges...)
	if f.UpdateGroupParticipantsFunc != nil {
		return f.UpdateGroupParticipantsFunc(ctx, jid, participantChanges, action)
	}
	return nil, nil
}

func (f *FakeClient) Disconnect() { f.DisconnectCalled = true }

func (f *FakeClient) IsConnected() bool {
	if f.IsConnectedFunc != nil {
		return f.IsConnectedFunc()
	}
	return true
}

func (f *FakeClient) Store() *store.Device {
	if f.StoreFunc != nil {
		return f.StoreFunc()
	}
	return nil
}

var _ bot.WhatsAppClient = (*FakeClient)(nil)
