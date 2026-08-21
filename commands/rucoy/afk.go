package rucoy

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
)

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
