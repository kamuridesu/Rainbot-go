package rucoy

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"github.com/kamuridesu/rainbot-go/internal/utils"
)

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
