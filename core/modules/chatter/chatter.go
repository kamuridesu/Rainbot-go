package chatter

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/kamuridesu/rainbot-go/core/commands"
	"github.com/kamuridesu/rainbot-go/core/services"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Chatter struct {
	Client *openai.Client
	Usage  *services.AIModelUsageService
}

var (
	SystemPrompt = ""
)

func GetSystemPrompt(botName string) string {
	if SystemPrompt != "" {
		return SystemPrompt
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Você é %s, um bot de WhatsApp. Responda sempre em português do Brasil. Para mencionar um usuario voce precisa incluir lid da pessoa que tem esse formato: @(\\d+)|(\\d+)@lid (nunca inclua @, é adicionado automaticamente pelo WhatsApp), caso não esteja incluido ignore a menção com @ e apenas diga o nome ou termos comuns para usuarios. Não precisa mencionar suas funcionalidades em todas as mensagens. Seja amigavel, seu publico alvo é composto principalmente de jovens e adolescentes, seja atual, mencione memes atuais, faça brincadeiras com brain rot, seja espontaneo. Suas respostas deverão sempre ser em minusculo, com um tom informal mas sem girias.\nComandos:\n", botName)
	for _, cmd := range *commands.GetLoadedCommands() {
		fmt.Fprintf(&b, "- %s [%s]: %s", cmd.Name, cmd.Category.Name, cmd.Description)
		if len(*cmd.Aliases) > 0 {
			fmt.Fprintf(&b, " | aliases: %s", strings.Join(*cmd.Aliases, ", "))
		}
		if len(*cmd.Examples) > 0 {
			fmt.Fprintf(&b, " | ex: %s", strings.Join(*cmd.Examples, "; "))
		}
		b.WriteByte('\n')
	}
	SystemPrompt = b.String()
	return SystemPrompt
}

func New(usage *services.AIModelUsageService) *Chatter {
	opts := []option.RequestOption{
		option.WithAPIKey(os.Getenv("AI_MODEL_KEY")),
	}
	openAiUrl := os.Getenv("AI_MODEL_URL")
	if openAiUrl != "" {
		slog.Info(fmt.Sprintf("Using %s AI Provider URL", openAiUrl))
		opts = append(opts, option.WithBaseURL(openAiUrl))
	}

	client := openai.NewClient(opts...)

	return &Chatter{
		Client: &client,
		Usage:  usage,
	}
}

func (c *Chatter) GenerateMessage(ctx context.Context, botName, message string) (string, error) {
	slog.Info("Loading Model...")
	model, err := AcquireModel(c.Usage)
	if err != nil {
		return "", err
	}
	slog.Info("Generating message")
	chatCompletion, err := c.Client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: model.Name,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(GetSystemPrompt(botName)),
			openai.UserMessage(message),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get AI response: %w", err)
	}

	if len(chatCompletion.Choices) == 0 {
		return "", fmt.Errorf("ai returned no completion choices")
	}

	slog.Info("Message generated")

	return chatCompletion.Choices[0].Message.Content, nil
}
