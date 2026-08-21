package commands

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/bot"
	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

var (
	testCmdOnce   sync.Once
	testCmdCalled bool
	testCategory  = NewCategory("test", nil)
)

const testCmdName = "commandtestonly"

func registerTestCommand() {
	testCmdOnce.Do(func() {
		if _, err := NewCommand(
			testCmdName,
			"a command used only by command_test.go",
			testCategory,
			&[]string{"commandtestalias"},
			&[]string{"${prefix}${alias} example"},
			true, false, false,
			func(m *messages.Message) { testCmdCalled = true },
		); err != nil {
			panic(err)
		}
	})
}

func newCommandsTestMessage(fake *botfakes.FakeClient, chat *models.Chat, command string) *messages.Message {
	args := []string{}
	return &messages.Message{
		Ctx:     context.Background(),
		Bot:     &bot.Bot{Client: fake, Name: strPtr("Rainbot")},
		Chat:    chat,
		Command: &command,
		Args:    &args,
		RawEvent: &events.Message{
			Info: types.MessageInfo{
				MessageSource: types.MessageSource{Chat: types.NewJID("123", types.GroupServer)},
				ID:            "stanza-1",
			},
		},
	}
}

func strPtr(s string) *string { return &s }

func TestFindCommandByNameAndAlias(t *testing.T) {
	cmd, err := FindCommand("help")
	if err != nil {
		t.Fatalf("FindCommand(help) error = %v", err)
	}
	if cmd.Name != "help" {
		t.Errorf("Name = %q, want %q", cmd.Name, "help")
	}

	if _, err := FindCommand("ajuda"); err != nil {
		t.Errorf("FindCommand(ajuda) (alias) error = %v", err)
	}
}

func TestFindCommandNotFound(t *testing.T) {
	if _, err := FindCommand("this-command-does-not-exist"); err == nil {
		t.Error("expected an error for an unregistered command")
	}
}

func TestNewCommandRejectsDuplicateName(t *testing.T) {
	registerTestCommand()

	defer func() {
		if recover() == nil {
			t.Error("expected NewCommand to panic on a duplicate name")
		}
	}()
	NewCommand(testCmdName, "dup", testCategory, nil, nil, false, false, false, func(m *messages.Message) {})
}

func TestNewCommandRejectsDuplicateCallback(t *testing.T) {
	registerTestCommand()
	existing, err := FindCommand(testCmdName)
	if err != nil {
		t.Fatalf("FindCommand() error = %v", err)
	}

	defer func() {
		if recover() == nil {
			t.Error("expected NewCommand to panic when reusing an already-registered callback")
		}
	}()
	NewCommand("a-different-name-entirely", "dup callback", testCategory, nil, nil, false, false, false, existing.Callable)
}

func TestGetCategoriesIncludesRegistered(t *testing.T) {
	registerTestCommand()
	categories := GetCategories()

	found := false
	for _, c := range categories {
		if c.Name == "test" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'test' among categories, got %v", categories)
	}
}

func TestGetCommandsFromCategory(t *testing.T) {
	registerTestCommand()
	cmds := *GetCommandsFromCategory("test")

	if len(cmds) != 1 || cmds[0].Name != testCmdName {
		t.Errorf("expected only %q in test, got %+v", testCmdName, cmds)
	}
}

func TestGetCommandsFromCategoryEmpty(t *testing.T) {
	cmds := *GetCommandsFromCategory("no-such-category")
	if len(cmds) != 0 {
		t.Errorf("expected no commands, got %+v", cmds)
	}
}

func TestFormatCommandHelp(t *testing.T) {
	registerTestCommand()
	cmd, err := FindCommand(testCmdName)
	if err != nil {
		t.Fatalf("FindCommand() error = %v", err)
	}

	got := formatCommandHelp(cmd, "!", testCmdName)
	if !strings.Contains(got, cmd.Description) {
		t.Errorf("expected the description in the help text, got %q", got)
	}
	if !strings.Contains(got, "commandtestalias") {
		t.Errorf("expected aliases listed, got %q", got)
	}
	if !strings.Contains(got, "!"+testCmdName+" example") {
		t.Errorf("expected the example with prefix/alias substituted, got %q", got)
	}
}

func TestDynamicMenuNoCategory(t *testing.T) {
	name := "Rainbot"
	menu, _ := dynamicMenu("", &bot.Bot{Name: &name})
	if !strings.Contains(menu, "Rainbot") {
		t.Errorf("expected the bot name in the menu, got %q", menu)
	}
	if !strings.Contains(menu, "misc") {
		t.Errorf("expected the 'misc' category (from the help command) listed, got %q", menu)
	}
}

func TestDynamicMenuWithCategory(t *testing.T) {
	registerTestCommand()
	name := "Rainbot"
	menu, _ := dynamicMenu("test", &bot.Bot{Name: &name})
	if !strings.Contains(menu, testCmdName) {
		t.Errorf("expected the command name listed, got %q", menu)
	}
}

func TestRunCommandExecutesCallback(t *testing.T) {
	registerTestCommand()
	testCmdCalled = false

	fake := &botfakes.FakeClient{}
	m := newCommandsTestMessage(fake, &models.Chat{AllowFun: 1}, testCmdName)

	RunCommand(m)

	if !testCmdCalled {
		t.Error("expected the command's callback to run")
	}
}

func TestRunCommandBlockedByChatCategorySetting(t *testing.T) {
	registerTestCommand()
	testCmdCalled = false

	fake := &botfakes.FakeClient{}
	m := newCommandsTestMessage(fake, &models.Chat{AllowFun: 0}, testCmdName)
	m.RawEvent.Info.IsGroup = true

	RunCommand(m)

	if testCmdCalled {
		t.Error("expected the callback NOT to run when fun commands are disabled for the chat")
	}
	if len(fake.SentMessages) == 0 {
		t.Error("expected a reply explaining the command is blocked")
	}
}

func TestRunCommandUnknownCommand(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newCommandsTestMessage(fake, &models.Chat{}, "this-command-does-not-exist")

	RunCommand(m)

	if len(fake.SentMessages) != 0 {
		t.Errorf("expected no reply for an unknown command, got %v", fake.SentMessages)
	}
}

func TestRunCommandAdminOnlyChatBlocksNonAdmin(t *testing.T) {
	fake := &botfakes.FakeClient{
		GetGroupInfoFunc: func(ctx context.Context, jid types.JID) (*types.GroupInfo, error) {
			return &types.GroupInfo{}, nil
		},
	}
	m := newCommandsTestMessage(fake, &models.Chat{AdminOnly: 1}, "help")
	m.RawEvent.Info.IsGroup = true
	m.Author = &models.Member{JID: "111@lid"}

	RunCommand(m)

	if len(fake.SentMessages) != 0 {
		t.Errorf("expected RunCommand to silently return for a non-admin in an admin-only chat, got %v", fake.SentMessages)
	}
}
