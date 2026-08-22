package rucoy

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"github.com/kamuridesu/rainbot-go/internal/utils"
)

func Uplevel(m *messages.Message) {
	args := *m.Args

	fromLevel, err := strconv.Atoi(args[0])
	if err != nil {
		m.Reply("level_atual precisa ser um numero. Exemplo: /uplevel 350 400 20kk", emojis.Fail)
		return
	}

	toLevel, err := strconv.Atoi(args[1])
	if err != nil {
		m.Reply("level_desejado precisa ser um numero. Exemplo: /uplevel 350 400 20kk", emojis.Fail)
		return
	}

	xpPerHour, err := parseRucoyXPPerHour(args[2])
	if err != nil {
		m.Reply("xp_por_hora precisa ser um numero. Exemplo: /uplevel 350 400 20kk", emojis.Fail)
		return
	}

	dailyHours, err := parseRucoyDailyHours(args[3:])
	if err != nil {
		m.Reply("horas_por_dia precisa ser um numero entre 1 e 24. Exemplo: /uplevel 350 400 20kk 8", emojis.Fail)
		return
	}

	if fromLevel <= 0 {
		m.Reply("O level atual precisa ser maior que zero.", emojis.Fail)
		return
	}
	if toLevel <= fromLevel {
		m.Reply("O level desejado precisa ser maior que o level atual.", emojis.Fail)
		return
	}
	if xpPerHour <= 0 {
		m.Reply("O XP/h precisa ser maior que zero.", emojis.Fail)
		return
	}

	params := url.Values{}
	params.Set("fromLevel", strconv.Itoa(fromLevel))
	params.Set("toLevel", strconv.Itoa(toLevel))

	requestURL := rucoyStatsApiURL + "/api/calculator/amount-exp?" + params.Encode()

	var result string
	err = utils.SendGETRequest(m.Ctx, http.DefaultClient, requestURL, &result, nil)
	if err != nil {
		m.Reply("Erro ao calcular uplevel: "+err.Error(), emojis.Fail)
		return
	}

	xpNeeded, err := strconv.ParseInt(strings.TrimSpace(result), 10, 64)
	if err != nil {
		m.Reply("Erro ao ler XP retornado pelo RucoyStats.", emojis.Fail)
		return
	}

	estimatedTime := formatRucoyDuration(xpNeeded, xpPerHour)
	reply := fmt.Sprintf(
		"Level: %d -> %d\nXP/h: %s\nXP faltando: %s\nTempo estimado: %s",
		fromLevel,
		toLevel,
		formatRucoyNumber(xpPerHour),
		formatRucoyNumber(xpNeeded),
		estimatedTime,
	)
	if dailyHours > 0 {
		reply += "\n" + formatRucoyDailyTrainingEstimate(estimatedTime, dailyHours)
	}

	m.Reply(reply, emojis.Success)
}

func parseRucoyXPPerHour(raw string) (int64, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, ",", ".")

	multiplier := float64(1)
	switch {
	case strings.HasSuffix(value, "kk"):
		multiplier = 1000000
		value = strings.TrimSuffix(value, "kk")
	case strings.HasSuffix(value, "m"):
		multiplier = 1000000
		value = strings.TrimSuffix(value, "m")
	case strings.HasSuffix(value, "k"):
		multiplier = 1000
		value = strings.TrimSuffix(value, "k")
	}

	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(number) || math.IsInf(number, 0) || number <= 0 {
		return 0, fmt.Errorf("invalid xp per hour")
	}

	return int64(number * multiplier), nil
}

func formatRucoyDuration(xpNeeded int64, xpPerHour int64) string {
	totalMinutes := int64(math.Ceil((float64(xpNeeded) / float64(xpPerHour)) * 60))
	hours := totalMinutes / 60
	minutes := totalMinutes % 60

	if hours == 0 {
		return fmt.Sprintf("%d minutos", minutes)
	}

	return fmt.Sprintf("%d horas e %d minutos", hours, minutes)
}
