package chatter

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/kamuridesu/rainbot-go/core/commands"
	"github.com/kamuridesu/rainbot-go/core/services"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Chatter struct {
	Client *openai.Client
	Usage  *services.AIModelUsageService
}

var promptCache sync.Map

const DefaultStyle = `- escreva tudo em minúsculo, tom informal e amigável
- seu público é principalmente jovem; faça piadas e referências a memes atuais quando combinar com a conversa, sem exagerar
- respostas curtas, como numa conversa de whatsapp
- só fale dos seus comandos quando perguntarem ou quando ajudar na resposta`

func GetSystemPrompt(botName, prefix, style string) string {
	if style == "" {
		style = DefaultStyle
	}
	return getBasePrompt(botName, prefix) + "\nestilo (siga sempre, sem ignorar as regras acima):\n" + style
}

func getBasePrompt(botName, prefix string) string {
	key := botName + "\x00" + prefix
	if p, ok := promptCache.Load(key); ok {
		return p.(string)
	}

	var b strings.Builder
	fmt.Fprintf(&b, `você é %[1]s, um bot de whatsapp. responda sempre em português do brasil.

prefixo deste chat: %[2]s
todo comando começa com esse prefixo (ex.: %[2]ssticker). sempre que citar um comando, escreva com o prefixo %[2]s.

sobre você:
- você foi criado por kamuri (github.com/kamuridesu). só mencione isso se perguntarem quem te criou, quem é seu dono ou de onde você veio.

menções:
- usuários aparecem no texto como @123456 ou 123456@lid
- para mencionar alguém, escreva só os números do lid com @lid (eg, 123456@lid), sem @ no inicio (o @ é adicionado automaticamente)
- se não tiver o lid da pessoa, não tente mencionar; chame pelo nome ou de um jeito genérico

comandos:
`, botName, prefix)

	for _, cmd := range *commands.GetLoadedCommands() {
		r := strings.NewReplacer("${prefix}", prefix, "${alias}", cmd.Name)
		fmt.Fprintf(&b, "- %s%s [%s]: %s", prefix, cmd.Name, cmd.Category.Name, r.Replace(cmd.Description))
		if cmd.Aliases != nil && len(*cmd.Aliases) > 0 {
			fmt.Fprintf(&b, " | aliases: %s%s", prefix, strings.Join(*cmd.Aliases, ", "+prefix))
		}
		if cmd.Examples != nil && len(*cmd.Examples) > 0 {
			ex := make([]string, len(*cmd.Examples))
			for i, e := range *cmd.Examples {
				ex[i] = r.Replace(e)
			}
			fmt.Fprintf(&b, " | ex: %s", strings.Join(ex, "; "))
		}
		b.WriteByte('\n')
	}

	p, _ := promptCache.LoadOrStore(key, b.String())
	return p.(string)
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

func (c *Chatter) GenerateMessage(ctx context.Context, botName, prefix, style, message string) (string, error) {
	slog.Info("Loading Model...")
	model, err := AcquireModel(c.Usage)
	if err != nil {
		return "", err
	}
	slog.Info("Generating message")
	chatCompletion, err := c.Client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: model.Name,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(GetSystemPrompt(botName, prefix, style)),
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
