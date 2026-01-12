package commands

import (
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strings"

	m "github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/bot"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
)

type Callback func(message *m.Message)

type Command struct {
	Name        string
	Aliases     *[]string
	Category    string
	Description string
	Examples    *[]string
	IsAdult     bool
	IsGame      bool
	IsFun       bool
	Callable    Callback
	Guards      []func(message *m.Message) error
}

type CommandList []*Command

var loadedCommands *CommandList = &CommandList{}

func init() {
	slog.Info("Resgistering meta-command help")
	help := func(msg *m.Message) {
		args := []string{}
		if msg.Args != nil {
			args = *msg.Args
		}
		commandOrCategory := strings.Join(args, " ")
		command, err := FindCommand(commandOrCategory)
		if err == nil {
			text := formatCommandHelp(command, msg.Chat.Prefix, commandOrCategory)
			msg.Reply(text)
		} else {
			menu := dynamicMenu(commandOrCategory, msg.Bot)
			msg.Reply(menu)
		}
	}

	NewCommand("help",
		"Mostra o menu de ajuda ou descrição de um comando",
		"misc",
		&[]string{"ajuda", "menu"},
		&[]string{"${prefix}${alias}", "${prefix}${alias} help"},
		false,
		false,
		false,
		help,
	)
}

func validateCommand(c *Command) error {

	for _, cmd := range *loadedCommands {

		if cmd.Name == c.Name || slices.Contains(*cmd.Aliases, c.Name) {
			return fmt.Errorf("Name %s is already reserved for a command", cmd.Name)
		}

		if reflect.ValueOf(cmd.Callable).Pointer() == reflect.ValueOf(c.Callable).Pointer() {
			return fmt.Errorf("Function assigned to cmd %s is already attached to %s", c.Name, cmd.Name)
		}

		for _, alias := range *c.Aliases {
			if alias == cmd.Name || slices.Contains(*cmd.Aliases, alias) {
				return fmt.Errorf("Alias %s is already present in function %s", alias, c.Name)
			}
		}

	}
	return nil
}

func NewCommand(name,
	desc,
	category string,
	aliases,
	examples *[]string,
	isFun,
	isGame,
	isAdult bool,
	callback Callback,
	guards ...func(message *m.Message) error) (*Command, error) {

	if aliases == nil {
		aliases = &[]string{}
	}
	if examples == nil {
		examples = &[]string{}
	}
	command := Command{
		Name:        name,
		Aliases:     aliases,
		Category:    category,
		Description: desc,
		Examples:    examples,
		IsAdult:     isAdult,
		IsGame:      isGame,
		IsFun:       isFun,
		Callable:    callback,
		Guards:      guards,
	}

	if err := validateCommand(&command); err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	*loadedCommands = append(*loadedCommands, &command)
	return &command, nil
}

func GetLoadedCommands() *CommandList {
	return loadedCommands
}

func GetCategories() *[]string {
	categories := []string{}
	for _, command := range *loadedCommands {
		if !slices.Contains(categories, command.Category) {
			categories = append(categories, command.Category)
		}
	}
	return &categories
}

func GetCommandsFromCategory(category string) *CommandList {
	var commands CommandList
	for _, command := range *loadedCommands {
		if command.Category == category {
			commands = append(commands, command)
		}
	}
	return &commands
}

func FindCommand(nameOrAlias string) (*Command, error) {
	for _, command := range *loadedCommands {
		if command.Name == nameOrAlias || slices.Contains(*command.Aliases, nameOrAlias) {
			return command, nil
		}
	}
	return nil, fmt.Errorf("no command found")
}

func formatMessage(text, term, prefix string) string {
	return strings.ReplaceAll(text, term, prefix)
}

func formatCommandHelp(command *Command, prefix string, commandOrCategory string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Comando: %s\n\n", commandOrCategory))
	sb.WriteString(fmt.Sprintf("Descrição: \n%s", command.Description))

	if command.Aliases != nil && len(*command.Aliases) > 0 {
		sb.WriteString("\n\nApelidos: \n- ")
		sb.WriteString(strings.Join(*command.Aliases, "\n- "))
	}

	if command.Examples != nil && len(*command.Examples) > 0 {
		joinedExamples := strings.Join(*command.Examples, "\n- ")
		formattedExamples := formatMessage(formatMessage(joinedExamples, "${prefix}", prefix), "${alias}", commandOrCategory)
		sb.WriteString("\n\nExemplos: \n- ")
		sb.WriteString(formattedExamples)
	}

	return sb.String()
}

func dynamicMenu(category string, bot *bot.Bot) string {
	if category == "" {
		return fmt.Sprintf("< %s > \n\n Categorias de comandos disponíveis: \n\n- %s",
			*bot.Name,
			strings.Join(*GetCategories(), "\n- "),
		)
	}

	commands := GetCommandsFromCategory(category)
	if commands == nil {
		return fmt.Sprintf("Categoria %s não encontrada!", category)
	}

	commandNames := []string{}

	for _, command := range *commands {
		commandNames = append(commandNames, command.Name)
	}

	return fmt.Sprintf("< %s > \n\n Comandos da categoria: \n\n- %s",
		*bot.Name,
		strings.Join(commandNames, "\n- "),
	)
}

func RunCommand(msg *m.Message) {
	slog.Info(fmt.Sprintf("Received command: %s", *msg.Command))

	if (msg.Chat.AdminOnly == 1) && (IsAdmin(msg) != nil) {
		return
	}

	cmd, err := FindCommand(*msg.Command)
	if err != nil {
		fmt.Println(err)
		return
	}

	isBlocked := (cmd.IsAdult && msg.Chat.AllowAdults == 0) ||
		(cmd.IsFun && msg.Chat.AllowFun == 0) ||
		(cmd.IsGame && msg.Chat.AllowGames == 0)

	if msg.IsFromGroup() && isBlocked {
		msg.Reply("Comando não pode ser usado neste grupo.", emojis.Fail)
		return
	}

	if cmd.Guards != nil {
		for _, guard := range cmd.Guards {
			if err := guard(msg); err != nil {
				slog.Error(err.Error())
				msg.Reply(err.Error())
				return
			}
		}
	}

	cmd.Callable(msg)
}
