package rucoy

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"github.com/kamuridesu/rainbot-go/internal/utils"
)

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
