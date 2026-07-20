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

type RucoyInactiveMember struct {
	Name        string
	DaysOffline int
}

type RucoyGoalMember struct {
	Name    string
	Level   int
	Missing int
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

const rucoyMinimumTrainingDurationSeconds = 8 * 60

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
	limit := 5
	if len(alternatives) < limit {
		limit = len(alternatives)
	}
	for i := 0; i < limit; i++ {
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
