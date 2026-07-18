package fun

import (
	"fmt"
	"html"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"github.com/kamuridesu/rainbot-go/internal/utils"
)

type RucoyGuildMember struct {
	Name          string
	Level         int
	Online        bool
	CharacterPath string
}

type ParsedRucoyGuildData struct {
	Guild   string
	Members []RucoyGuildMember
}

func (p *ParsedRucoyGuildData) String(onlineOnly bool) string {
	sb := strings.Builder{}
	if onlineOnly {
		fmt.Fprintf(&sb, "Online em %s:\n", p.Guild)
	} else {
		fmt.Fprintf(&sb, "Membros em %s:\n", p.Guild)
	}

	onlineCount := 0
	for _, member := range p.Members {
		if !onlineOnly || (onlineOnly && member.Online) {
			fmt.Fprintf(&sb, "- %s: lv %d\n", member.Name, member.Level)
			if onlineOnly && member.Online {
				onlineCount++
			}
		}
	}
	if onlineOnly && onlineCount == 0 {
		return fmt.Sprintf("Nenhum jogador online em %s.", p.Guild)
	}
	return sb.String()
}

func parseRucoyResponse(data string, guildName string) *ParsedRucoyGuildData {
	parsedData := &ParsedRucoyGuildData{
		Guild:   guildName,
		Members: make([]RucoyGuildMember, 0),
	}

	rowRegex := regexp.MustCompile(`(?s)<tr>\s*<td>\s*<a href="(/characters/[^"]+)">([^<]+)</a>(.*?)</tr>`)

	levelRegex := regexp.MustCompile(`<td>\s*(\d+)\s*</td>`)

	matches := rowRegex.FindAllStringSubmatch(data, -1)
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}

		characterPath := strings.TrimSpace(html.UnescapeString(match[1]))
		name := strings.TrimSpace(html.UnescapeString(match[2]))

		restOfRow := match[3]

		isOnline := strings.Contains(restOfRow, ">Online</span>")

		level := 0
		levelMatch := levelRegex.FindStringSubmatch(restOfRow)
		if len(levelMatch) >= 2 {
			level, _ = strconv.Atoi(levelMatch[1])
		}

		parsedData.Members = append(parsedData.Members, RucoyGuildMember{
			Name:          name,
			Level:         level,
			Online:        isOnline,
			CharacterPath: characterPath,
		})
	}

	return parsedData
}

func RucoyOnlineGuild(m *messages.Message) {
	guild := strings.Join(*m.Args, " ")

	url := fmt.Sprintf("https://www.rucoyonline.com/guild/%s", url.PathEscape(guild))
	var response string
	err := utils.SendGETRequest(m.Ctx, http.DefaultClient, url, &response, nil)
	if err != nil {
		m.Reply("Erro ao ler dados da guilda: "+err.Error(), emojis.Fail)
		return
	}

	rucoyGuild := parseRucoyResponse(response, guild)
	if len(rucoyGuild.Members) == 0 {
		m.Reply("Guild não encontrada", emojis.Fail)
		return
	}

	m.Reply(rucoyGuild.String(true), emojis.Success)
}

func RucoyMenu(m *messages.Message) {
	m.Reply(`Comandos Rucoy

/online Nome-da-Guild
Mostra quem esta online na guild.

/upskill skillatual skilldesejada tickrate
Calcula o tempo para upar skill.
Exemplo: /upskill 366 400 42000

/uplevel levelatual leveldesejado xp/h
Calcula o tempo para upar level.
Exemplo: /uplevel 350 400 20kk

/afk Nome-da-Guild
Mostra jogadores com 7 dias ou mais sem logar.

/meta LEVEL Nome-da-Guild
Mostra quem ainda nao bateu a meta.
Exemplo: /meta 400 Nome-da-Guild`, emojis.Success)
}

type RucoyInactiveMember struct {
	Name        string
	DaysOffline int
}

type RucoyGoalMember struct {
	Name    string
	Level   int
	Missing int
}

func RucoyAFKGuild(m *messages.Message) {
	guild := strings.Join(*m.Args, " ")

	requestURL := fmt.Sprintf("https://www.rucoyonline.com/guild/%s", url.PathEscape(guild))
	var response string
	err := utils.SendGETRequest(m.Ctx, http.DefaultClient, requestURL, &response, nil)
	if err != nil {
		m.Reply("Erro ao ler dados da guilda: "+err.Error(), emojis.Fail)
		return
	}

	rucoyGuild := parseRucoyResponse(response, guild)
	if len(rucoyGuild.Members) == 0 {
		m.Reply("Guild não encontrada", emojis.Fail)
		return
	}

	inactiveMembers := make([]RucoyInactiveMember, 0)
	for index, member := range rucoyGuild.Members {
		if index > 0 {
			time.Sleep(1500 * time.Millisecond)
		}

		lastOnline, err := fetchRucoyLastOnlineDays(m, member)
		if err != nil {
			m.Reply(fmt.Sprintf("Erro ao ler perfil de %s: %s", member.Name, err.Error()), emojis.Fail)
			return
		}

		if lastOnline >= 7 {
			inactiveMembers = append(inactiveMembers, RucoyInactiveMember{
				Name:        member.Name,
				DaysOffline: lastOnline,
			})
		}
	}

	if len(inactiveMembers) == 0 {
		m.Reply(fmt.Sprintf("Nenhum jogador inativo em %s.", rucoyGuild.Guild), emojis.Success)
		return
	}

	sort.SliceStable(inactiveMembers, func(i, j int) bool {
		return inactiveMembers[i].DaysOffline > inactiveMembers[j].DaysOffline
	})

	sb := strings.Builder{}
	fmt.Fprintf(&sb, "Jogadores inativos em %s:\n\n", rucoyGuild.Guild)
	for _, member := range inactiveMembers {
		fmt.Fprintf(&sb, "%s %d dias offline\n", member.Name, member.DaysOffline)
	}

	m.Reply(sb.String(), emojis.Success)
}

func RucoyMetaGuild(m *messages.Message) {
	args := *m.Args

	goal, err := strconv.Atoi(args[0])
	if err != nil || goal <= 0 {
		m.Reply("Use: /meta 400 NOME DA GUILD", emojis.Fail)
		return
	}

	guild := strings.Join(args[1:], " ")
	requestURL := fmt.Sprintf("https://www.rucoyonline.com/guild/%s", url.PathEscape(guild))
	var response string
	err = utils.SendGETRequest(m.Ctx, http.DefaultClient, requestURL, &response, nil)
	if err != nil {
		m.Reply("Erro ao ler dados da guilda: "+err.Error(), emojis.Fail)
		return
	}

	rucoyGuild := parseRucoyResponse(response, guild)
	if len(rucoyGuild.Members) == 0 {
		m.Reply("Guild nÃ£o encontrada", emojis.Fail)
		return
	}

	membersBelowGoal := make([]RucoyGoalMember, 0)
	for _, member := range rucoyGuild.Members {
		if member.Level < goal {
			membersBelowGoal = append(membersBelowGoal, RucoyGoalMember{
				Name:    member.Name,
				Level:   member.Level,
				Missing: goal - member.Level,
			})
		}
	}

	if len(membersBelowGoal) == 0 {
		m.Reply(fmt.Sprintf("Todos os membros de %s jÃ¡ bateram a meta %d.", rucoyGuild.Guild, goal), emojis.Success)
		return
	}

	sort.SliceStable(membersBelowGoal, func(i, j int) bool {
		if membersBelowGoal[i].Missing == membersBelowGoal[j].Missing {
			return membersBelowGoal[i].Name < membersBelowGoal[j].Name
		}
		return membersBelowGoal[i].Missing < membersBelowGoal[j].Missing
	})

	sb := strings.Builder{}
	fmt.Fprintf(&sb, "Meta level %d em %s:\n\n", goal, rucoyGuild.Guild)
	for _, member := range membersBelowGoal {
		fmt.Fprintf(&sb, "%s - level %d - falta %d\n", member.Name, member.Level, member.Missing)
	}

	m.Reply(sb.String(), emojis.Success)
}

func fetchRucoyLastOnlineDays(m *messages.Message, member RucoyGuildMember) (int, error) {
	if member.CharacterPath == "" {
		return 0, fmt.Errorf("link do personagem não encontrado")
	}

	requestURL := "https://www.rucoyonline.com" + member.CharacterPath
	response, err := sendRucoyGETWithRetry(m, requestURL)
	if err != nil {
		return 0, err
	}

	return parseRucoyLastOnlineDays(response), nil
}

func sendRucoyGETWithRetry(m *messages.Message, requestURL string) (string, error) {
	delays := []time.Duration{
		3 * time.Second,
		7 * time.Second,
		15 * time.Second,
	}

	for attempt := 0; attempt <= len(delays); attempt++ {
		req, err := http.NewRequestWithContext(m.Ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return "", fmt.Errorf("failed to build request to %s: %v", requestURL, err)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to send request to %s: %v", requestURL, err)
		}

		resBody, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return "", fmt.Errorf("failed to read body from %s: %v", requestURL, readErr)
		}

		if res.StatusCode == http.StatusTooManyRequests {
			if attempt == len(delays) {
				return "", fmt.Errorf("site do Rucoy limitou muitas requisicoes, tente novamente em alguns minutos")
			}

			time.Sleep(delays[attempt])
			continue
		}

		if res.StatusCode > 400 {
			return "", fmt.Errorf("error : status is %d and body is %s", res.StatusCode, string(resBody))
		}

		return string(resBody), nil
	}

	return "", fmt.Errorf("site do Rucoy limitou muitas requisicoes, tente novamente em alguns minutos")
}

func parseRucoyLastOnlineDays(data string) int {
	lastOnlineRegex := regexp.MustCompile(`(?is)<td>\s*Last online\s*</td>\s*<td>\s*([^<]+)\s*</td>`)
	match := lastOnlineRegex.FindStringSubmatch(data)
	if len(match) < 2 {
		return 0
	}

	lastOnline := strings.ToLower(strings.TrimSpace(html.UnescapeString(match[1])))
	lastOnline = strings.Join(strings.Fields(lastOnline), " ")
	if lastOnline == "" || strings.Contains(lastOnline, "online") {
		return 0
	}

	parts := strings.Fields(lastOnline)
	if len(parts) < 2 {
		return 0
	}

	value, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}

	unit := parts[1]
	switch {
	case strings.HasPrefix(unit, "minute"), strings.HasPrefix(unit, "hour"):
		return 0
	case strings.HasPrefix(unit, "day"):
		return value
	case strings.HasPrefix(unit, "week"):
		return value * 7
	case strings.HasPrefix(unit, "month"):
		return value * 30
	case strings.HasPrefix(unit, "year"):
		return value * 365
	default:
		return 0
	}
}

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

	requestURL := "https://rucoystatsapi.net/api/calculator/amount-time?" + params.Encode()

	var result string
	err = utils.SendGETRequest(m.Ctx, http.DefaultClient, requestURL, &result, nil)
	if err != nil {
		m.Reply("Erro ao calcular upskill: "+err.Error(), emojis.Fail)
		return
	}

	m.Reply(fmt.Sprintf(
		"Upskill Rucoy\n\nSkill atual: %d\nSkill desejada: %d\nTickrate: %d\nTempo estimado: %s",
		fromSkill,
		toSkill,
		tickrate,
		formatUpskillTime(result),
	), emojis.Success)
}

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

	requestURL := "https://rucoystatsapi.net/api/calculator/amount-exp?" + params.Encode()

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

	m.Reply(fmt.Sprintf(
		"Uplevel Rucoy\n\nLevel: %d -> %d\nXP/h: %s\nXP faltando: %s\nTempo estimado: %s",
		fromLevel,
		toLevel,
		formatRucoyNumber(xpPerHour),
		formatRucoyNumber(xpNeeded),
		formatRucoyDuration(xpNeeded, xpPerHour),
	), emojis.Success)
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

func formatRucoyNumber(value int64) string {
	raw := strconv.FormatInt(value, 10)
	parts := make([]string, 0, len(raw)/3+1)

	for len(raw) > 3 {
		parts = append([]string{raw[len(raw)-3:]}, parts...)
		raw = raw[:len(raw)-3]
	}

	parts = append([]string{raw}, parts...)
	return strings.Join(parts, ".")
}
