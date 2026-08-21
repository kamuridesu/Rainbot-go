package rucoy

import (
	"html"
	"regexp"
	"strconv"
	"strings"
)

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
