package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kamuridesu/rainbot-go/internal/database/models"
	"github.com/kamuridesu/rainbot-go/internal/services"
)

func validateBool(value string, index int) error {
	if value != "sim" && value != "nao" && value != "não" {
		return fmt.Errorf("falha ao processar valor da config, valores aceitos na linha %d são apenas sim/nao/não", index)
	}
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func ParseSetupText(args []string, chat *models.Chat, chatService *services.ChatService) error {
	for index, line := range args {
		data := strings.Split(line, "=")
		if len(data) != 2 {
			return fmt.Errorf("falha ao processar configuração na linha %d", index)
		}
		key := data[0]
		value := strings.ToLower(data[1])
		switch key {
		case "limiteDeAvisos":
			int_value, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("falha ao processar valor da config na linha %d, apenas numeros sao aceitos depois do =", index)
			}
			chat.WarnBanThreshold = int_value
		case "apenasAdmin":
			if err := validateBool(value, index); err != nil {
				return err
			}
			chat.AdminOnly = boolToInt(value == "sim")
		case "prefixo":
			if len(value) != 1 {
				return fmt.Errorf("o tamanho do prefixo deve ser de apenas 1 caractere")
			}
			chat.Prefix = value
		case "ativarBot":
			if err := validateBool(value, index); err != nil {
				return err
			}
			chat.IsBotEnabled = boolToInt(value == "sim")
		case "palavrasProibidas":
			chat.CustomProfanityWords = value
		case "filtroDeProfanidade":
			if err := validateBool(value, index); err != nil {
				return err
			}
			chat.ProfanityFilterEnabled = boolToInt(value == "sim")
		case "boasVindas":
			chat.WelcomeMessage = value
		case "contarMensagens":
			if err := validateBool(value, index); err != nil {
				return err
			}
			chat.CountMessages = boolToInt(value == "sim")
		case "ativarGames":
			if err := validateBool(value, index); err != nil {
				return err
			}
			chat.AllowGames = boolToInt(value == "sim")
		case "ativarDiversao":
			if err := validateBool(value, index); err != nil {
				return err
			}
			chat.AllowFun = boolToInt(value == "sim")
		case "AtivarAdultos":
			if err := validateBool(value, index); err != nil {
				return err
			}
			chat.AllowAdults = boolToInt(value == "sim")
		}
	}

	return chatService.UpdateChat(chat)

}

func boolTHuman(b bool) string {
	if b {
		return "sim"
	}
	return "não"
}

func GetHumanReadableSetup(chat *models.Chat) string {

	message := fmt.Sprintf("prefixo=%s\n", chat.Prefix)
	message += fmt.Sprintf("ativarBot=%s\n", boolTHuman(chat.IsBotEnabled == 1))
	message += fmt.Sprintf("apenasAdmin=%s\n", boolTHuman(chat.AdminOnly == 1))
	message += fmt.Sprintf("limiteDeAvisos=%d\n", chat.WarnBanThreshold)
	message += fmt.Sprintf("ativarGames=%s\n", boolTHuman(chat.AllowGames == 1))
	message += fmt.Sprintf("ativarAdultos=%s\n", boolTHuman(chat.AllowAdults == 1))
	message += fmt.Sprintf("ativarDiversao=%s\n", boolTHuman(chat.AllowFun == 1))
	message += fmt.Sprintf("contarMensagens=%s\n", boolTHuman(chat.CountMessages == 1))
	message += fmt.Sprintf("filtroDeProfanidade=%s\n", boolTHuman(chat.ProfanityFilterEnabled == 1))
	message += fmt.Sprintf("palavrasProibidas=%s\n", chat.CustomProfanityWords)
	message += fmt.Sprintf("boasVindas=%s", chat.WelcomeMessage)

	return message

}
