package fun

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"slices"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func Casal(m *messages.Message) {

	info, err := m.Bot.Client.GetGroupInfo(m.RawEvent.Info.Chat)
	if err != nil {
		slog.Error(err.Error())
		m.Reply("houve um erro ao processar: "+err.Error(), emojis.Fail)
		return
	}
	totalMembers := len(info.Participants)

	var members []types.JID

	for range 2 {
		for {
			mem := info.Participants[rand.IntN(totalMembers)].LID.ToNonAD()
			if mem.String() != m.Bot.Client.Store.LID.ToNonAD().String() && !slices.Contains(members, mem) {
				members = append(members, mem)
				break
			}
		}
	}

	message := fmt.Sprintf("❤️❤️ Meu casal ❤️❤️\n\n @%s + @%s", members[0].User, members[1].User)

	m.SendMessage(&waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			ContextInfo: &waE2E.ContextInfo{
				MentionedJID: []string{members[0].String(), members[1].String()},
			},
			Text: proto.String(message),
		},
	}, m.RawEvent.Info.Chat)

	m.React(emojis.Success)

}
