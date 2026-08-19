package fun

import (
	"encoding/json"
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

// var rather than const so tests can point these at a local httptest server.
var (
	rucoyBaseURL     = "https://www.rucoyonline.com"
	rucoyStatsApiURL = "https://rucoystatsapi.net"
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

	url := fmt.Sprintf("%s/guild/%s", rucoyBaseURL, url.PathEscape(guild))
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

type RucoyInactiveMember struct {
	Name        string
	DaysOffline int
}

type RucoyGoalMember struct {
	Name    string
	Level   int
	Missing int
}

type RucoyCharacterInfo struct {
	Name       string
	Level      int
	Guild      string
	Title      string
	LastOnline string
}

type RucoyStatsGuildResponse struct {
	Players []RucoyStatsGuildPlayer `json:"players"`
}

type RucoyStatsGuildPlayer struct {
	Name       string `json:"name"`
	LastOnline string `json:"lastOnline"`
}

type RucoyLevelTableEntry struct {
	Level              int   `json:"level"`
	ExpLoss            int64 `json:"expLoss"`
	GoldBlackOneNeeded int64 `json:"goldBlackOneNeeded"`
}

type RucoyTrainingMonster struct {
	Name       string
	Defense    int
	HP         float64
	Powertrain bool
}

type RucoyTrainingResult struct {
	Mode                  string
	Monster               string
	Efficiency            float64
	DurationSeconds       float64
	MinimumDuration       float64
	MaxDamage             int
	MaxCriticalDamage     int
	NextMonster           string
	RequiredStats         int
	StatsNeededFor1Damage int
	BestShortMonster      string
	BestShortEfficiency   float64
	BestShortDuration     float64
}

type RucoyTrainingAlternative struct {
	Attack          int
	Monster         string
	Efficiency      float64
	DurationSeconds float64
}

type RucoyUpskillOptions struct {
	DailyHours   int
	Vocation     string
	ManaPerSkill int64
}

const rucoyMinimumTrainingDurationSeconds = 8 * 60
const rucoyAFKProfileDelay = 2500 * time.Millisecond
const rucoyUltimateManaPotionMin = 600
const rucoyUltimateManaPotionMax = 900
const rucoyUltimateManaPotionPackSize = 200
const rucoyUltimateManaPotionPackGold = 130000

func RucoyAFKGuild(m *messages.Message) {
	guild := strings.Join(*m.Args, " ")

	inactiveMembers, found, err := fetchRucoyStatsInactiveMembers(m, guild)
	if err == nil && found {
		m.Reply(formatRucoyAFKReply(guild, inactiveMembers, 0, false), emojis.Success)
		return
	}

	requestURL := fmt.Sprintf("%s/guild/%s", rucoyBaseURL, url.PathEscape(guild))
	response, err := sendRucoyGETWithRetry(m, requestURL)
	if err != nil {
		m.Reply("Erro ao ler dados da guilda: "+err.Error(), emojis.Fail)
		return
	}

	rucoyGuild := parseRucoyResponse(response, guild)
	if len(rucoyGuild.Members) == 0 {
		m.Reply("Guild não encontrada", emojis.Fail)
		return
	}

	inactiveMembers = make([]RucoyInactiveMember, 0)
	failedMembers := make([]string, 0)
	checkedProfiles := 0
	for _, member := range rucoyGuild.Members {
		if member.Online {
			continue
		}

		if checkedProfiles > 0 {
			time.Sleep(rucoyAFKProfileDelay)
		}
		checkedProfiles++

		lastOnline, err := fetchRucoyLastOnlineDays(m, member)
		if err != nil {
			failedMembers = append(failedMembers, member.Name)
			continue
		}

		if lastOnline >= 7 {
			inactiveMembers = append(inactiveMembers, RucoyInactiveMember{
				Name:        member.Name,
				DaysOffline: lastOnline,
			})
		}
	}

	m.Reply(formatRucoyAFKReply(rucoyGuild.Guild, inactiveMembers, len(failedMembers), checkedProfiles > 0 && len(failedMembers) == checkedProfiles), emojis.Success)
}

func fetchRucoyStatsInactiveMembers(m *messages.Message, guild string) ([]RucoyInactiveMember, bool, error) {
	requestURL := fmt.Sprintf("%s/api/Guild/onlines-of-guild=%s", rucoyStatsApiURL, url.PathEscape(guild))
	response, err := sendRucoyGETWithRetry(m, requestURL)
	if err != nil {
		return nil, false, err
	}

	response = strings.TrimSpace(response)
	if response == "" {
		return nil, false, nil
	}

	var guildResponse RucoyStatsGuildResponse
	if err := json.Unmarshal([]byte(response), &guildResponse); err != nil {
		return nil, false, err
	}
	if len(guildResponse.Players) == 0 {
		return nil, false, nil
	}

	inactiveMembers := make([]RucoyInactiveMember, 0)
	for _, player := range guildResponse.Players {
		daysOffline := parseRucoyLastOnlineText(player.LastOnline)
		if daysOffline >= 7 {
			inactiveMembers = append(inactiveMembers, RucoyInactiveMember{
				Name:        player.Name,
				DaysOffline: daysOffline,
			})
		}
	}

	return inactiveMembers, true, nil
}

func formatRucoyAFKReply(guild string, inactiveMembers []RucoyInactiveMember, failedProfiles int, allProfilesFailed bool) string {
	sort.SliceStable(inactiveMembers, func(i, j int) bool {
		return inactiveMembers[i].DaysOffline > inactiveMembers[j].DaysOffline
	})

	sb := strings.Builder{}
	if len(inactiveMembers) > 0 {
		fmt.Fprintf(&sb, "Jogadores inativos em %s:\n\n", guild)
		for _, member := range inactiveMembers {
			fmt.Fprintf(&sb, "%s %d dias offline\n", member.Name, member.DaysOffline)
		}
	} else if allProfilesFailed {
		fmt.Fprintf(&sb, "Nao consegui verificar os perfis de %s agora.", guild)
	} else if failedProfiles > 0 {
		fmt.Fprintf(&sb, "Nenhum jogador inativo encontrado nos perfis verificados em %s.", guild)
	} else {
		fmt.Fprintf(&sb, "Nenhum jogador inativo em %s.", guild)
	}

	if failedProfiles > 0 {
		fmt.Fprintf(&sb, "\n\nNao consegui verificar %d perfil(is) agora por limite do site. Tente novamente em alguns minutos.", failedProfiles)
	}

	return sb.String()
}

func RucoyMetaGuild(m *messages.Message) {
	args := *m.Args

	goal, err := strconv.Atoi(args[0])
	if err != nil || goal <= 0 {
		m.Reply("Use: /meta 400 NOME DA GUILD", emojis.Fail)
		return
	}

	guild := strings.Join(args[1:], " ")
	requestURL := fmt.Sprintf("%s/guild/%s", rucoyBaseURL, url.PathEscape(guild))
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

func RucoyInfo(m *messages.Message) {
	player := strings.Join(*m.Args, " ")
	requestURL := fmt.Sprintf("%s/characters/%s", rucoyBaseURL, url.PathEscape(player))

	response, err := sendRucoyGETWithRetry(m, requestURL)
	if err != nil {
		if strings.Contains(err.Error(), "status is 404") {
			m.Reply("Jogador nao encontrado", emojis.Fail)
			return
		}
		m.Reply("Erro ao ler dados do jogador: "+err.Error(), emojis.Fail)
		return
	}

	info := parseRucoyCharacterInfo(response)
	if info.Name == "" || info.Level == 0 {
		m.Reply("Jogador nao encontrado", emojis.Fail)
		return
	}

	if info.Guild == "" {
		info.Guild = "-"
	}
	if info.Title == "" {
		info.Title = "-"
	}

	levelTableEntry, err := fetchRucoyLevelTableEntry(m, info.Level)
	if err != nil {
		m.Reply("Erro ao ler tabela de XP do RucoyStats: "+err.Error(), emojis.Fail)
		return
	}

	statusLabel := "Ultima vez online"
	statusValue := info.LastOnline
	if statusValue == "" {
		statusValue = "-"
	} else if strings.Contains(strings.ToLower(statusValue), "online") {
		statusLabel = "Status"
		statusValue = "Online"
	}

	m.Reply(fmt.Sprintf(
		"Info Rucoy\n\nNome: %s\nLevel: %d\nGuild: %s\nTitulo: %s\n%s: %s\nBlack Skull: %s gold\nXP Mobwin: %s XP",
		info.Name,
		info.Level,
		info.Guild,
		info.Title,
		statusLabel,
		statusValue,
		formatRucoyNumber(levelTableEntry.GoldBlackOneNeeded),
		formatRucoyNumber(absRucoyNumber(levelTableEntry.ExpLoss)),
	), emojis.Success)
}

func fetchRucoyLevelTableEntry(m *messages.Message, level int) (RucoyLevelTableEntry, error) {
	var levelTable []RucoyLevelTableEntry
	err := utils.SendGETRequest(m.Ctx, http.DefaultClient, rucoyStatsApiURL+"/api/calculator/experiences", &levelTable, nil)
	if err != nil {
		return RucoyLevelTableEntry{}, err
	}

	for _, entry := range levelTable {
		if entry.Level == level {
			return entry, nil
		}
	}

	return RucoyLevelTableEntry{}, fmt.Errorf("level %d nao encontrado", level)
}

func fetchRucoyLastOnlineDays(m *messages.Message, member RucoyGuildMember) (int, error) {
	if member.CharacterPath == "" {
		return 0, fmt.Errorf("link do personagem não encontrado")
	}

	requestURL := rucoyBaseURL + member.CharacterPath
	response, err := sendRucoyGETWithRetry(m, requestURL)
	if err != nil {
		return 0, err
	}

	return parseRucoyLastOnlineDays(response), nil
}

func sendRucoyGETWithRetry(m *messages.Message, requestURL string) (string, error) {
	delays := []time.Duration{
		5 * time.Second,
		15 * time.Second,
		30 * time.Second,
		60 * time.Second,
	}

	for attempt := 0; attempt <= len(delays); attempt++ {
		req, err := http.NewRequestWithContext(m.Ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return "", fmt.Errorf("failed to build request to %s: %v", requestURL, err)
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Rainbot-go Rucoy commands)")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9,pt-BR;q=0.8,pt;q=0.7")

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

			time.Sleep(rucoyRetryDelay(res, delays[attempt]))
			continue
		}

		if res.StatusCode >= 400 {
			return "", fmt.Errorf("error : status is %d and body is %s", res.StatusCode, string(resBody))
		}

		return string(resBody), nil
	}

	return "", fmt.Errorf("site do Rucoy limitou muitas requisicoes, tente novamente em alguns minutos")
}

func rucoyRetryDelay(res *http.Response, fallback time.Duration) time.Duration {
	retryAfter := strings.TrimSpace(res.Header.Get("Retry-After"))
	if retryAfter == "" {
		return fallback
	}

	seconds, err := strconv.Atoi(retryAfter)
	if err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	retryAt, err := http.ParseTime(retryAfter)
	if err != nil {
		return fallback
	}

	delay := time.Until(retryAt)
	if delay <= 0 {
		return fallback
	}
	return delay
}

func parseRucoyCharacterInfo(data string) RucoyCharacterInfo {
	info := RucoyCharacterInfo{}
	rowRegex := regexp.MustCompile(`(?is)<tr>\s*<td>\s*([^<]+?)\s*</td>\s*<td>\s*(.*?)\s*</td>\s*</tr>`)

	for _, match := range rowRegex.FindAllStringSubmatch(data, -1) {
		if len(match) < 3 {
			continue
		}

		field := strings.ToLower(stripRucoyHTML(match[1]))
		value := stripRucoyHTML(match[2])

		switch field {
		case "name":
			info.Name = value
		case "level":
			info.Level, _ = strconv.Atoi(value)
		case "guild":
			info.Guild = value
		case "title":
			info.Title = value
		case "last online":
			info.LastOnline = value
		}
	}

	return info
}

func stripRucoyHTML(value string) string {
	tagRegex := regexp.MustCompile(`(?s)<[^>]+>`)
	value = tagRegex.ReplaceAllString(value, " ")
	value = html.UnescapeString(value)
	return strings.Join(strings.Fields(value), " ")
}

func parseRucoyLastOnlineDays(data string) int {
	lastOnlineRegex := regexp.MustCompile(`(?is)<td>\s*Last online\s*</td>\s*<td>\s*([^<]+)\s*</td>`)
	match := lastOnlineRegex.FindStringSubmatch(data)
	if len(match) < 2 {
		return 0
	}

	return parseRucoyLastOnlineText(match[1])
}

func parseRucoyLastOnlineText(value string) int {
	lastOnline := strings.ToLower(strings.TrimSpace(stripRucoyHTML(value)))
	lastOnline = strings.Join(strings.Fields(lastOnline), " ")
	if lastOnline == "" || lastOnline == "online" || strings.Contains(lastOnline, "currently online") || strings.Contains(lastOnline, "loading") {
		return 0
	}

	parts := strings.Fields(lastOnline)
	if len(parts) < 2 {
		return 0
	}

	valueIndex := -1
	offlineValue := 0
	for index, part := range parts {
		parsedValue, err := strconv.Atoi(part)
		if err == nil {
			valueIndex = index
			offlineValue = parsedValue
			break
		}
		if part == "a" || part == "an" {
			valueIndex = index
			offlineValue = 1
			break
		}
	}
	if valueIndex < 0 || valueIndex+1 >= len(parts) {
		return 0
	}

	unit := parts[valueIndex+1]
	switch {
	case strings.HasPrefix(unit, "minute"), strings.HasPrefix(unit, "hour"):
		return 0
	case strings.HasPrefix(unit, "day"):
		return offlineValue
	case strings.HasPrefix(unit, "week"):
		return offlineValue * 7
	case strings.HasPrefix(unit, "month"):
		return offlineValue * 30
	case strings.HasPrefix(unit, "year"):
		return offlineValue * 365
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
	if options.Vocation != "" {
		reply += "\n\n" + formatRucoyUpskillManaEstimate(estimatedTime, tickrate, options)
	}

	m.Reply(reply, emojis.Success)
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
		"Uplevel Rucoy\n\nLevel: %d -> %d\nXP/h: %s\nXP faltando: %s\nTempo estimado: %s",
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

func RucoyTrain(m *messages.Message) {
	args := *m.Args

	attack, err := strconv.Atoi(args[0])
	if err != nil {
		m.Reply("arma precisa ser um numero. Exemplo: /train 5 351 391 -50", emojis.Fail)
		return
	}

	baseLevel, err := strconv.Atoi(args[1])
	if err != nil {
		m.Reply("level precisa ser um numero. Exemplo: /train 5 351 391 -50", emojis.Fail)
		return
	}

	statLevel, err := strconv.Atoi(args[2])
	if err != nil {
		m.Reply("skill precisa ser um numero. Exemplo: /train 5 351 391 -50", emojis.Fail)
		return
	}

	extraStats, err := strconv.Atoi(args[3])
	if err != nil {
		m.Reply("add precisa ser um numero. Exemplo: /train 5 351 391 -50", emojis.Fail)
		return
	}

	targetEfficiency := 90.0
	if len(args) >= 5 {
		targetEfficiency, err = strconv.ParseFloat(strings.ReplaceAll(args[4], ",", "."), 64)
		if err != nil {
			m.Reply("eficiencia precisa ser um numero. Exemplo: /train 5 351 391 -50 90", emojis.Fail)
			return
		}
	}

	if err := validateRucoyTrainInput(attack, baseLevel, statLevel, extraStats, targetEfficiency); err != nil {
		m.Reply(err.Error(), emojis.Fail)
		return
	}

	afkResult := calculateRucoyTraining(baseLevel, statLevel, extraStats, attack, false, targetEfficiency)
	powerResult := calculateRucoyTraining(baseLevel, statLevel, extraStats, attack, true, targetEfficiency)

	sb := strings.Builder{}
	fmt.Fprintf(&sb, "Calculadora Train Rucoy\n\n")
	fmt.Fprintf(&sb, "Setup: arma %d | level %d | skill %d | add %+d\n", attack, baseLevel, statLevel, extraStats)
	fmt.Fprintf(&sb, "Skill efetiva: %d\n", statLevel+extraStats)
	fmt.Fprintf(&sb, "Eficiencia alvo: %.0f%%+\n\n", targetEfficiency)
	writeRucoyTrainingResult(&sb, afkResult, targetEfficiency)
	if afkResult.Monster == "" {
		writeRucoyTrainingWeaponAlternatives(&sb, attack, baseLevel, statLevel, extraStats, targetEfficiency, false)
	}
	sb.WriteString("\n")
	writeRucoyTrainingResult(&sb, powerResult, targetEfficiency)
	if powerResult.Monster == "" {
		writeRucoyTrainingWeaponAlternatives(&sb, attack, baseLevel, statLevel, extraStats, targetEfficiency, true)
	}

	m.Reply(sb.String(), emojis.Success)
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

func formatRucoyUpskillManaEstimate(estimatedTime string, tickrate int, options RucoyUpskillOptions) string {
	totalMinutes, ok := parseRucoyFormattedDurationMinutes(estimatedTime)
	if !ok || tickrate <= 0 || options.ManaPerSkill <= 0 {
		return ""
	}

	totalTicks := ceilDivInt64(int64(tickrate)*totalMinutes, 60)
	totalMana := totalTicks * options.ManaPerSkill
	minPotions := ceilDivInt64(totalMana, rucoyUltimateManaPotionMax)
	maxPotions := ceilDivInt64(totalMana, rucoyUltimateManaPotionMin)
	minPacks := ceilDivInt64(minPotions, rucoyUltimateManaPotionPackSize)
	maxPacks := ceilDivInt64(maxPotions, rucoyUltimateManaPotionPackSize)
	minCost := minPacks * rucoyUltimateManaPotionPackGold
	maxCost := maxPacks * rucoyUltimateManaPotionPackGold

	return fmt.Sprintf(
		"Gasto estimado com Ultimate Mana Potion\nClasse: %s\nMana total: %s\nPotions: %s a %s\nPacks de 200: %s a %s\nCusto: %s a %s gold",
		options.Vocation,
		formatRucoyNumber(totalMana),
		formatRucoyNumber(minPotions),
		formatRucoyNumber(maxPotions),
		formatRucoyNumber(minPacks),
		formatRucoyNumber(maxPacks),
		formatRucoyNumber(minCost),
		formatRucoyNumber(maxCost),
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

func validateRucoyTrainInput(attack int, baseLevel int, statLevel int, extraStats int, targetEfficiency float64) error {
	if baseLevel < 1 || baseLevel > 1000 {
		return fmt.Errorf("level precisa estar entre 1 e 1000. Exemplo: /train 5 351 391 -50")
	}
	if statLevel < 5 || statLevel > 1000 {
		return fmt.Errorf("skill precisa estar entre 5 e 1000. Exemplo: /train 5 351 391 -50")
	}
	if extraStats < -80 || extraStats > 126 {
		return fmt.Errorf("add precisa estar entre -80 e 126. Exemplo: /train 5 351 391 -50")
	}
	if attack < 4 || attack > 60 || attack == 6 || attack == 8 || attack == 10 || attack == 12 || attack == 14 {
		return fmt.Errorf("arma invalida. Use o ataque da arma de treino. Exemplo: /train 5 351 391 -50")
	}
	if targetEfficiency < 35 || targetEfficiency > 99 {
		return fmt.Errorf("eficiencia precisa estar entre 35 e 99. Exemplo: /train 5 351 391 -50 90")
	}

	return nil
}

func calculateRucoyTraining(baseLevel int, statLevel int, extraStats int, attack int, powertrain bool, targetEfficiency float64) RucoyTrainingResult {
	crit := 0.01
	critMulti := 1.05
	totalStat := statLevel + extraStats
	ticks := 10.0
	specMulti := 1.0
	mode := "AFK Train"
	if powertrain {
		ticks = 38
		specMulti = 1.5
		mode = "Powertrain"
	}

	minDamage := specMulti * (float64(baseLevel)/4 + float64(attack*totalStat)/20)
	maxDamage := specMulti * (float64(baseLevel)/4 + float64(attack*totalStat)/10)
	avgCritMulti := 1 + (critMulti-1)/2
	targetProb := 1 - math.Pow((100-targetEfficiency)/100, 1/ticks)

	result := RucoyTrainingResult{
		Mode:            mode,
		MinimumDuration: rucoyMinimumTrainingDurationSeconds,
	}

	for _, monster := range rucoyTrainingMonsters() {
		if powertrain && !monster.Powertrain {
			continue
		}

		prob := math.Min((1-crit)*(maxDamage-float64(monster.Defense))/(maxDamage-minDamage)+crit, 1)
		if targetProb < prob {
			finalProb := 100 - 100*math.Pow(1-prob, ticks)
			duration := rucoyTrainingDuration(monster, minDamage, maxDamage, crit, critMulti, avgCritMulti, prob)
			if duration <= 0 {
				continue
			}
			if duration < rucoyMinimumTrainingDurationSeconds {
				if result.BestShortDuration < duration {
					result.BestShortMonster = monster.Name
					result.BestShortEfficiency = finalProb
					result.BestShortDuration = duration
				}
				continue
			}

			if result.DurationSeconds < duration {
				result.Monster = monster.Name
				result.Efficiency = finalProb
				result.DurationSeconds = duration
				result.MaxDamage = int(math.Floor(maxDamage)) - monster.Defense
				result.MaxCriticalDamage = int(math.Floor(maxDamage*critMulti)) - monster.Defense
			} else if result.DurationSeconds == duration {
				result.Monster += " & " + monster.Name
			}
			continue
		}

		result.NextMonster = monster.Name
		result.RequiredStats = int(math.Ceil(
			(20*float64(monster.Defense)-20*float64(baseLevel)/4*specMulti)/
				(float64(attack)*specMulti*(2-(targetProb-crit)/(1-crit))),
		)) - totalStat
		result.StatsNeededFor1Damage = int(math.Ceil(
			10*((float64(1+monster.Defense)/specMulti)-float64(baseLevel)/4)/float64(attack),
		)) - totalStat
		break
	}

	return result
}

func rucoyTrainingDuration(monster RucoyTrainingMonster, minDamage float64, maxDamage float64, crit float64, critMulti float64, avgCritMulti float64, prob float64) float64 {
	monsterDefense := float64(monster.Defense)
	var damagePerSecond float64
	if minDamage < monsterDefense {
		damagePerSecond = crit*(maxDamage*avgCritMulti-monsterDefense) +
			(1-crit)*(maxDamage-monsterDefense)*prob/2
	} else {
		damagePerSecond = crit*(maxDamage*avgCritMulti-monsterDefense) +
			(1-crit)*(maxDamage+minDamage-2*monsterDefense)/2
	}

	if damagePerSecond <= 0 {
		return 0
	}

	return monster.HP / damagePerSecond
}

func writeRucoyTrainingResult(sb *strings.Builder, result RucoyTrainingResult, targetEfficiency float64) {
	fmt.Fprintf(sb, "%s:\n", result.Mode)
	if result.Monster == "" {
		fmt.Fprintf(sb, "Nenhum monstro viavel com %s+ e %.0f%%+ de eficiencia.\n", formatRucoyTrainingDuration(result.MinimumDuration), targetEfficiency)
		if result.BestShortMonster != "" {
			fmt.Fprintf(sb, "Melhor acima da eficiencia, mas ruim: %s\n", result.BestShortMonster)
			fmt.Fprintf(sb, "Ele morreria em media em %s.\n", formatRucoyTrainingDuration(result.BestShortDuration))
		}
		writeRucoyTrainingNextStep(sb, result, targetEfficiency)
		return
	}

	fmt.Fprintf(sb, "Melhor local: %s\n", result.Monster)
	fmt.Fprintf(sb, "Eficiencia estimada: %.1f%%\n", result.Efficiency)
	fmt.Fprintf(sb, "Tempo medio ate matar o mob: %s\n", formatRucoyTrainingDuration(result.DurationSeconds))
	if result.DurationSeconds > 450 {
		sb.WriteString("Obs: o mob pode exaurir antes, por volta de 07:30.\n")
	}
	fmt.Fprintf(sb, "Dano max: %d | crit max: %d\n", result.MaxDamage, result.MaxCriticalDamage)

	writeRucoyTrainingNextStep(sb, result, targetEfficiency)
}

func writeRucoyTrainingNextStep(sb *strings.Builder, result RucoyTrainingResult, targetEfficiency float64) {
	if result.NextMonster == "" {
		sb.WriteString("Proximo mob: nenhum acima na tabela atual.\n")
		return
	}

	fmt.Fprintf(sb, "Proximo mob: %s\n", result.NextMonster)
	if result.RequiredStats > 0 {
		fmt.Fprintf(sb, "Para avancar: falta +%d skill/add para %.0f%%+ de eficiencia.\n", result.RequiredStats, targetEfficiency)
		return
	}
	if result.StatsNeededFor1Damage > 0 {
		fmt.Fprintf(sb, "Para avancar: falta +%d skill/add para dar 1 dano max.\n", result.StatsNeededFor1Damage)
		return
	}

	sb.WriteString("Para avancar: voce ja esta perto; teste uma eficiencia alvo menor se quiser forcar esse mob.\n")
}

func writeRucoyTrainingWeaponAlternatives(sb *strings.Builder, currentAttack int, baseLevel int, statLevel int, extraStats int, targetEfficiency float64, powertrain bool) {
	alternatives := rucoyTrainingWeaponAlternatives(currentAttack, baseLevel, statLevel, extraStats, targetEfficiency, powertrain)
	mode := "AFK Train"
	if powertrain {
		mode = "Powertrain"
	}
	if len(alternatives) == 0 {
		fmt.Fprintf(sb, "Sugestoes: nem mudando so a arma de treino achei um %s 08:00+ nessa tabela.\n", mode)
		return
	}

	sb.WriteString("Sugestoes com arma de treino:\n")
	limit := min(len(alternatives), 5)
	for i := range limit {
		alternative := alternatives[i]
		fmt.Fprintf(
			sb,
			"- arma %d: %s por %s (%.1f%%)\n",
			alternative.Attack,
			alternative.Monster,
			formatRucoyTrainingDuration(alternative.DurationSeconds),
			alternative.Efficiency,
		)
	}
}

func rucoyTrainingWeaponAlternatives(currentAttack int, baseLevel int, statLevel int, extraStats int, targetEfficiency float64, powertrain bool) []RucoyTrainingAlternative {
	alternatives := make([]RucoyTrainingAlternative, 0)
	for _, attack := range rucoyTrainingWeaponAttacks() {
		if attack == currentAttack {
			continue
		}

		result := calculateRucoyTraining(baseLevel, statLevel, extraStats, attack, powertrain, targetEfficiency)
		if result.Monster == "" {
			continue
		}

		alternatives = append(alternatives, RucoyTrainingAlternative{
			Attack:          attack,
			Monster:         result.Monster,
			Efficiency:      result.Efficiency,
			DurationSeconds: result.DurationSeconds,
		})
	}

	sort.SliceStable(alternatives, func(i, j int) bool {
		return alternatives[i].DurationSeconds > alternatives[j].DurationSeconds
	})

	return alternatives
}

func formatRucoyTrainingDuration(seconds float64) string {
	totalSeconds := int(math.Round(seconds))
	minutes := totalSeconds / 60
	remainingSeconds := totalSeconds % 60

	return fmt.Sprintf("%02d:%02d", minutes, remainingSeconds)
}

func rucoyTrainingWeaponAttacks() []int {
	return []int{4, 5, 7, 9, 11, 13}
}

func rucoyTrainingMonsters() []RucoyTrainingMonster {
	return []RucoyTrainingMonster{
		{Name: "Rat Lv.1", Defense: 4, HP: 25, Powertrain: false},
		{Name: "Rat Lv.3", Defense: 7, HP: 35, Powertrain: false},
		{Name: "Crow Lv.6", Defense: 13, HP: 40, Powertrain: false},
		{Name: "Wolf Lv.9", Defense: 17, HP: 50, Powertrain: false},
		{Name: "Scorpion Lv.12", Defense: 18, HP: 50, Powertrain: false},
		{Name: "Cobra Lv.13", Defense: 18, HP: 50, Powertrain: false},
		{Name: "Worm Lv.14", Defense: 19, HP: 55, Powertrain: false},
		{Name: "Goblin Lv.15", Defense: 21, HP: 60, Powertrain: true},
		{Name: "Mummy Lv.25", Defense: 36, HP: 80, Powertrain: true},
		{Name: "Pharaoh Lv.35", Defense: 51, HP: 100, Powertrain: true},
		{Name: "Assassin Lv.45", Defense: 71, HP: 120, Powertrain: true},
		{Name: "Assassin Lv.50", Defense: 81, HP: 140, Powertrain: true},
		{Name: "Assassin Ninja Lv.55", Defense: 91, HP: 160, Powertrain: true},
		{Name: "Skeleton Archer Lv.80", Defense: 101, HP: 300, Powertrain: false},
		{Name: "Zombie Lv.65", Defense: 106, HP: 200, Powertrain: true},
		{Name: "Skeleton Lv.75", Defense: 121, HP: 300, Powertrain: true},
		{Name: "Skeleton Warrior Lv.90", Defense: 146, HP: 375, Powertrain: true},
		{Name: "Vampire Lv.100", Defense: 171, HP: 450, Powertrain: true},
		{Name: "Vampire Lv.110", Defense: 186, HP: 530, Powertrain: true},
		{Name: "Drow Ranger Lv.125", Defense: 191, HP: 600, Powertrain: false},
		{Name: "Drow Mage Lv.130", Defense: 191, HP: 600, Powertrain: false},
		{Name: "Drow Assassin Lv.120", Defense: 221, HP: 620, Powertrain: true},
		{Name: "Drow Sorceress Lv.140", Defense: 221, HP: 600, Powertrain: false},
		{Name: "Drow Fighter Lv.135", Defense: 246, HP: 680, Powertrain: true},
		{Name: "Lizard Archer Lv.160", Defense: 271, HP: 650, Powertrain: false},
		{Name: "Lizard Shaman Lv.170", Defense: 276, HP: 600, Powertrain: false},
		{Name: "Dead Eyes Lv.170", Defense: 276, HP: 600, Powertrain: false},
		{Name: "Lizard Warrior Lv.150", Defense: 301, HP: 680, Powertrain: true},
		{Name: "Djinn Lv.150", Defense: 301, HP: 640, Powertrain: true},
		{Name: "Lizard High Shaman Lv.190", Defense: 326, HP: 740, Powertrain: false},
		{Name: "Gargoyle Lv.190", Defense: 326, HP: 740, Powertrain: true},
		{Name: "Dragon Hatchling Lv.240", Defense: 331, HP: 10000, Powertrain: false},
		{Name: "Lizard Captain Lv.180", Defense: 361, HP: 815, Powertrain: true},
		{Name: "Dragon Lv.250", Defense: 501, HP: 20000, Powertrain: false},
		{Name: "Minotaur Lv.225", Defense: 511, HP: 4250, Powertrain: true},
		{Name: "Minotaur Lv.250", Defense: 601, HP: 5000, Powertrain: true},
		{Name: "Dragon Warden Lv.280", Defense: 626, HP: 30000, Powertrain: false},
		{Name: "Ice Elemental Lv.300", Defense: 676, HP: 40000, Powertrain: false},
		{Name: "Minotaur Lv.275", Defense: 691, HP: 5750, Powertrain: true},
		{Name: "Ice Dragon Lv.320", Defense: 726, HP: 45000, Powertrain: false},
		{Name: "Yeti Lv.350", Defense: 826, HP: 55000, Powertrain: false},
		{Name: "Lava Golem Lv.375", Defense: 900, HP: 65000, Powertrain: false},
		{Name: "Orthrus Lv.400", Defense: 1300, HP: 75000, Powertrain: false},
		{Name: "Demon Lv.450", Defense: 1550, HP: 100000, Powertrain: false},
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

func absRucoyNumber(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
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
