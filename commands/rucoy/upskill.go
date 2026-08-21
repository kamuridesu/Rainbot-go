package rucoy

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"github.com/kamuridesu/rainbot-go/internal/utils"
)

func Upskill(m *messages.Message) {
	args := *m.Args

	fromSkill, err := strconv.Atoi(args[0])
	if err != nil {
		m.Reply("skillatual precisa ser um numero. Exemplo: /upskill 400 450 5000", emojis.Fail)
		return
	}

	toSkill, err := strconv.Atoi(args[1])
	if err != nil {
		m.Reply("skilldesejada precisa ser um numero. Exemplo: /upskill 400 450 5000", emojis.Fail)
		return
	}

	tickrate, err := strconv.Atoi(args[2])
	if err != nil {
		m.Reply("tickrate precisa ser um numero. Exemplo: /upskill 400 450 5000", emojis.Fail)
		return
	}

	options, err := parseRucoyUpskillOptions(args[3:])
	if err != nil {
		m.Reply("Opcional invalido. Use horas de 1 a 24 e/ou classe kina, pally ou mage. Exemplos: /upskill 400 450 5000 8 ou /upskill 400 450 5000 kina 8", emojis.Fail)
		return
	}

	if fromSkill < 55 {
		fromSkill = 55
	}
	if toSkill <= fromSkill {
		m.Reply("A skill desejada precisa ser maior que a skill atual.", emojis.Fail)
		return
	}
	if tickrate < 200 {
		tickrate = 200
	}
	if tickrate > 50400 {
		tickrate = 50400
	}

	params := url.Values{}
	params.Set("fromValue", strconv.Itoa(fromSkill))
	params.Set("toLevel", strconv.Itoa(toSkill))
	params.Set("trainMode", strconv.Itoa(tickrate))

	requestURL := rucoyStatsApiURL + "/api/calculator/amount-time?" + params.Encode()

	var result string
	err = utils.SendGETRequest(m.Ctx, http.DefaultClient, requestURL, &result, nil)
	if err != nil {
		m.Reply("Erro ao calcular upskill: "+err.Error(), emojis.Fail)
		return
	}

	estimatedTime := formatUpskillTime(result)
	reply := fmt.Sprintf(
		"Upskill Rucoy\n\nSkill atual: %d\nSkill desejada: %d\nTickrate: %d\nTempo estimado: %s",
		fromSkill,
		toSkill,
		tickrate,
		estimatedTime,
	)
	if options.DailyHours > 0 {
		reply += "\n" + formatRucoyDailyTrainingEstimate(estimatedTime, options.DailyHours)
	}

	var manaEstimate RucoyUpskillManaEstimate
	if options.Vocation != "" {
		var ok bool
		manaEstimate, ok = calculateRucoyUpskillManaEstimate(estimatedTime, tickrate, options)
		if ok {
			reply += "\n\n" + formatRucoyUpskillManaEstimate(options, manaEstimate)
		}
	}

	if options.Vocation != "" && manaEstimate.TotalMana > 0 {
		card, err := generateRucoyUpskillCard(RucoyUpskillCardData{
			FromSkill:     fromSkill,
			ToSkill:       toSkill,
			EstimatedTime: estimatedTime,
			DailyHours:    options.DailyHours,
			Options:       options,
			ManaEstimate:  manaEstimate,
		})
		if err != nil {
			slog.Error("failed to generate Rucoy upskill card", "error", err)
			m.Reply(reply+"\n\nNao consegui gerar a imagem do /upskill.", emojis.Fail)
			return
		}
		if _, err := m.ReplyMedia(card, reply, messages.ImageMessage, emojis.Success); err != nil {
			slog.Error("failed to send Rucoy upskill card", "error", err)
			m.Reply(reply+"\n\nNao consegui enviar a imagem do /upskill.", emojis.Fail)
		}
		return
	}

	m.Reply(reply, emojis.Success)
}

func formatUpskillTime(raw string) string {
	parts := strings.Split(strings.TrimSpace(raw), ":")

	switch len(parts) {
	case 3:
		days, _ := strconv.Atoi(parts[0])
		hours, _ := strconv.Atoi(parts[1])
		minutes, _ := strconv.Atoi(parts[2])
		return fmt.Sprintf("%d horas e %d minutos", days*24+hours, minutes)
	case 2:
		hours, _ := strconv.Atoi(parts[0])
		minutes, _ := strconv.Atoi(parts[1])
		return fmt.Sprintf("%d horas e %d minutos", hours, minutes)
	case 1:
		minutes, _ := strconv.Atoi(parts[0])
		return fmt.Sprintf("%d minutos", minutes)
	default:
		return raw
	}
}

func parseRucoyDailyHours(args []string) (int, error) {
	if len(args) == 0 {
		return 0, nil
	}

	hours, err := strconv.Atoi(args[0])
	if err != nil || hours < 1 || hours > 24 {
		return 0, fmt.Errorf("invalid daily hours")
	}

	return hours, nil
}

func parseRucoyUpskillOptions(args []string) (RucoyUpskillOptions, error) {
	options := RucoyUpskillOptions{}
	for _, arg := range args {
		arg = strings.ToLower(strings.TrimSpace(arg))
		if arg == "" {
			continue
		}

		if hours, err := strconv.Atoi(arg); err == nil {
			if options.DailyHours != 0 || hours < 1 || hours > 24 {
				return RucoyUpskillOptions{}, fmt.Errorf("invalid daily hours")
			}
			options.DailyHours = hours
			continue
		}

		vocation, manaPerSkill, ok := parseRucoyUpskillVocation(arg)
		if !ok || options.Vocation != "" {
			return RucoyUpskillOptions{}, fmt.Errorf("invalid vocation")
		}
		options.Vocation = vocation
		options.ManaPerSkill = manaPerSkill
	}

	return options, nil
}

func parseRucoyUpskillVocation(value string) (string, int64, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "kina", "knight":
		return "Kina", 50, true
	case "pally", "paladin":
		return "Pally", 50, true
	case "mage", "mago":
		return "Mage", 40, true
	default:
		return "", 0, false
	}
}

func formatRucoyDailyTrainingEstimate(estimatedTime string, dailyHours int) string {
	totalMinutes, ok := parseRucoyFormattedDurationMinutes(estimatedTime)
	if !ok || dailyHours <= 0 {
		return ""
	}

	dailyMinutes := int64(dailyHours * 60)
	days := totalMinutes / dailyMinutes
	remainingMinutes := totalMinutes % dailyMinutes
	remainingHours := remainingMinutes / 60
	minutes := remainingMinutes % 60

	if days == 0 {
		if remainingHours == 0 {
			return fmt.Sprintf("Treinando %dh por dia: %d minutos", dailyHours, minutes)
		}
		return fmt.Sprintf("Treinando %dh por dia: %d horas e %d minutos", dailyHours, remainingHours, minutes)
	}

	if remainingHours == 0 && minutes == 0 {
		return fmt.Sprintf("Treinando %dh por dia: %d dias", dailyHours, days)
	}

	return fmt.Sprintf("Treinando %dh por dia: %d dias, %d horas e %d minutos", dailyHours, days, remainingHours, minutes)
}

func calculateRucoyUpskillManaEstimate(estimatedTime string, tickrate int, options RucoyUpskillOptions) (RucoyUpskillManaEstimate, bool) {
	totalMinutes, ok := parseRucoyFormattedDurationMinutes(estimatedTime)
	if !ok || tickrate <= 0 || options.ManaPerSkill <= 0 {
		return RucoyUpskillManaEstimate{}, false
	}

	totalSkillUses := totalMinutes * 60
	totalMana := totalSkillUses * options.ManaPerSkill
	minPotions := ceilDivInt64(totalMana, rucoyUltimateManaPotionMax)
	maxPotions := ceilDivInt64(totalMana, rucoyUltimateManaPotionMin)
	minPacks := ceilDivInt64(minPotions, rucoyUltimateManaPotionPackSize)
	maxPacks := ceilDivInt64(maxPotions, rucoyUltimateManaPotionPackSize)
	minCost := minPacks * rucoyUltimateManaPotionPackGold
	maxCost := maxPacks * rucoyUltimateManaPotionPackGold
	totalArrows := int64(0)
	arrowCost := int64(0)
	if options.Vocation == "Pally" {
		totalArrows = totalSkillUses * rucoyPallySkillsPerSecond * rucoyPallyArrowsPerSkill
		arrowCost = ceilDivInt64(totalArrows, rucoyPallyArrowBundleSize) * rucoyPallyArrowBundleGold
		minCost += arrowCost
		maxCost += arrowCost
	}

	return RucoyUpskillManaEstimate{
		TotalMana:   totalMana,
		MinPotions:  minPotions,
		MaxPotions:  maxPotions,
		TotalArrows: totalArrows,
		ArrowCost:   arrowCost,
		MinCost:     minCost,
		MaxCost:     maxCost,
	}, true
}

func formatRucoyUpskillManaEstimate(options RucoyUpskillOptions, estimate RucoyUpskillManaEstimate) string {
	arrowText := ""
	costLabel := "Custo"
	if estimate.TotalArrows > 0 {
		arrowText = fmt.Sprintf("\nFlechas: %s\nCusto flechas: %s gold", formatRucoyNumber(estimate.TotalArrows), formatRucoyNumber(estimate.ArrowCost))
		costLabel = "Custo total"
	}

	return fmt.Sprintf(
		"Gasto estimado com Ultimate Mana Potion\nClasse: %s\nMana total: %s\nPotions: %s a %s%s\n%s: %s a %s gold",
		options.Vocation,
		formatRucoyNumber(estimate.TotalMana),
		formatRucoyNumber(estimate.MinPotions),
		formatRucoyNumber(estimate.MaxPotions),
		arrowText,
		costLabel,
		formatRucoyNumber(estimate.MinCost),
		formatRucoyNumber(estimate.MaxCost),
	)
}

func ceilDivInt64(value int64, divisor int64) int64 {
	if divisor <= 0 {
		return 0
	}
	if value <= 0 {
		return 0
	}
	return (value + divisor - 1) / divisor
}

func parseRucoyFormattedDurationMinutes(value string) (int64, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}

	hoursRegex := regexp.MustCompile(`^(\d+)\s+horas?\s+e\s+(\d+)\s+minutos?$`)
	if match := hoursRegex.FindStringSubmatch(value); len(match) == 3 {
		hours, _ := strconv.ParseInt(match[1], 10, 64)
		minutes, _ := strconv.ParseInt(match[2], 10, 64)
		return hours*60 + minutes, true
	}

	minutesRegex := regexp.MustCompile(`^(\d+)\s+minutos?$`)
	if match := minutesRegex.FindStringSubmatch(value); len(match) == 2 {
		minutes, _ := strconv.ParseInt(match[1], 10, 64)
		return minutes, true
	}

	return 0, false
}
