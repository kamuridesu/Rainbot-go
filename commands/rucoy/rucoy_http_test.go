package rucoy

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
)

func withRucoyServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	origBase, origStats := rucoyBaseURL, rucoyStatsApiURL
	rucoyBaseURL, rucoyStatsApiURL = server.URL, server.URL
	t.Cleanup(func() {
		rucoyBaseURL, rucoyStatsApiURL = origBase, origStats
	})

	return server
}

func guildRowHTML(name string, level int, online bool) string {
	status := "Offline"
	if online {
		status = "Online"
	}
	return `<tr><td><a href="/characters/` + name + `">` + name + `</a></td><td>` + strconv.Itoa(level) + `</td><td><span class="status">` + status + `</span></td></tr>`
}

func lastReplyText(fake *botfakes.FakeClient) string {
	for index := len(fake.SentMessages) - 1; index >= 0; index-- {
		if imageMessage := fake.SentMessages[index].Message.GetImageMessage(); imageMessage != nil {
			return imageMessage.GetCaption()
		}
		if textMessage := fake.SentMessages[index].Message.GetExtendedTextMessage(); textMessage != nil {
			return textMessage.GetText()
		}
	}
	return ""
}

func sentTextCount(fake *botfakes.FakeClient) int {
	total := 0
	for _, sent := range fake.SentMessages {
		if sent.Message.GetExtendedTextMessage() != nil {
			total++
		}
	}
	return total
}

func sentImageCount(fake *botfakes.FakeClient) int {
	total := 0
	for _, sent := range fake.SentMessages {
		if sent.Message.GetImageMessage() != nil {
			total++
		}
	}
	return total
}

func TestRucoyOnlineGuildSuccess(t *testing.T) {
	withRucoyServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<table>" + guildRowHTML("John", 50, true) + guildRowHTML("Jane", 30, false) + "</table>"))
	})

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"TestGuild"}

	RucoyOnlineGuild(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + reply), got %d", len(fake.SentMessages))
	}
	text := lastReplyText(fake)
	if !strings.Contains(text, "John") || strings.Contains(text, "Jane") {
		t.Errorf("expected only the online member listed, got %q", text)
	}
}

func TestRucoyOnlineGuildNotFound(t *testing.T) {
	withRucoyServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body>no guild here</body></html>"))
	})

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"NoSuchGuild"}

	RucoyOnlineGuild(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "não encontrada") {
		t.Errorf("expected a not-found reply, got %q", text)
	}
}

func TestRucoyOnlineGuildHTTPError(t *testing.T) {
	withRucoyServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"TestGuild"}

	RucoyOnlineGuild(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "Erro ao ler dados") {
		t.Errorf("expected an error reply, got %q", text)
	}
}

func TestRucoyMetaGuildBelowGoal(t *testing.T) {
	withRucoyServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<table>" + guildRowHTML("John", 350, true) + guildRowHTML("Jane", 450, false) + "</table>"))
	})

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"400", "TestGuild"}

	RucoyMetaGuild(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "John") || strings.Contains(text, "Jane") {
		t.Errorf("expected only the member below the goal listed, got %q", text)
	}
}

func TestRucoyMetaGuildInvalidGoal(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"notanumber", "TestGuild"}

	RucoyMetaGuild(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "Use:") {
		t.Errorf("expected a usage reply, got %q", text)
	}
}

func TestRucoyInfoSuccess(t *testing.T) {
	withRucoyServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/characters/"):
			w.Write([]byte(`<table>
				<tr><td>Name</td><td>John</td></tr>
				<tr><td>Level</td><td>123</td></tr>
				<tr><td>Guild</td><td>TestGuild</td></tr>
				<tr><td>Title</td><td>Hero</td></tr>
				<tr><td>Last Online</td><td>Online</td></tr>
			</table>`))
		case strings.Contains(r.URL.Path, "/api/calculator/experiences"):
			w.Write([]byte(`[{"level":123,"expLoss":-1000,"goldBlackOneNeeded":50000}]`))
		}
	})

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"John"}

	RucoyInfo(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + reply), got %d", len(fake.SentMessages))
	}
	text := lastReplyText(fake)
	if !strings.Contains(text, "John") || !strings.Contains(text, "123") || !strings.Contains(text, "TestGuild") {
		t.Errorf("expected the reply to include player info, got %q", text)
	}
}

func TestRucoyInfoNotFound(t *testing.T) {
	withRucoyServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"NoSuchPlayer"}

	RucoyInfo(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "nao encontrado") {
		t.Errorf("expected a not-found reply, got %q", text)
	}
}

func TestRucoyAFKGuildDetectsInactiveMember(t *testing.T) {
	withRucoyServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/guild/"):
			w.Write([]byte("<table>" + guildRowHTML("John", 50, false) + "</table>"))
		case strings.Contains(r.URL.Path, "/characters/"):
			w.Write([]byte(`<td>Last online</td><td>10 days ago</td>`))
		}
	})

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"TestGuild"}

	RucoyAFKGuild(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + reply), got %d", len(fake.SentMessages))
	}
	text := lastReplyText(fake)
	if !strings.Contains(text, "John") || !strings.Contains(text, "10 dias") {
		t.Errorf("expected the inactive member to be reported, got %q", text)
	}
}

func TestRucoyAFKGuildUsesRucoyStatsLastOnline(t *testing.T) {
	characterRequests := 0
	withRucoyServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/api/Guild/onlines-of-guild="):
			w.Write([]byte(`{"players":[
				{"name":"Old Player","lastOnline":"30 days ago"},
				{"name":"Recent Player","lastOnline":"6 days ago"},
				{"name":"Year Player","lastOnline":"over 1 year ago"},
				{"name":"Online Player","lastOnline":"currently online"}
			]}`))
		case strings.Contains(r.URL.Path, "/characters/"):
			characterRequests++
			w.Write([]byte(`<td>Last online</td><td>10 days ago</td>`))
		}
	})

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"TestGuild"}

	RucoyAFKGuild(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "Old Player 30 dias offline") ||
		!strings.Contains(text, "Year Player 365 dias offline") ||
		strings.Contains(text, "Recent Player") ||
		strings.Contains(text, "Online Player") {
		t.Errorf("expected inactive members from RucoyStats reply, got %q", text)
	}
	if characterRequests != 0 {
		t.Errorf("expected no character profile requests when RucoyStats succeeds, got %d", characterRequests)
	}
}

func TestRucoyAFKGuildFallsBackWhenRucoyStatsIsUnavailable(t *testing.T) {
	characterRequests := 0
	withRucoyServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/api/Guild/onlines-of-guild="):
			w.WriteHeader(http.StatusInternalServerError)
		case strings.Contains(r.URL.Path, "/guild/"):
			w.Write([]byte("<table>" + guildRowHTML("Online Guy", 50, true) + guildRowHTML("Offline Guy", 50, false) + "</table>"))
		case strings.Contains(r.URL.Path, "/characters/"):
			characterRequests++
			w.Write([]byte(`<td>Last online</td><td>10 days ago</td>`))
		}
	})

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"TestGuild"}

	RucoyAFKGuild(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "Offline Guy 10 dias offline") || strings.Contains(text, "Online Guy") {
		t.Errorf("expected fallback to check only offline members, got %q", text)
	}
	if characterRequests != 1 {
		t.Errorf("expected only one character profile request in fallback, got %d", characterRequests)
	}
}

func TestUpskillSuccess(t *testing.T) {
	withRucoyServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("02:30:00"))
	})

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"400", "450", "5000"}

	Upskill(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "Upskill Rucoy") || !strings.Contains(text, "horas") {
		t.Errorf("expected a formatted upskill reply, got %q", text)
	}
}

func TestUpskillWithDailyHours(t *testing.T) {
	withRucoyServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("01:02:30"))
	})

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"400", "450", "5000", "8"}

	Upskill(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "Tempo estimado: 26 horas e 30 minutos") ||
		!strings.Contains(text, "Treinando 8h por dia: 3 dias, 2 horas e 30 minutos") {
		t.Errorf("expected daily training estimate in upskill reply, got %q", text)
	}
}

func TestUpskillWithManaEstimate(t *testing.T) {
	withRucoyServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("01:02:30"))
	})

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"400", "450", "5000", "kina"}

	Upskill(m)

	text := lastReplyText(fake)
	expectedParts := []string{
		"Tempo estimado: 26 horas e 30 minutos",
		"Gasto estimado com Ultimate Mana Potion",
		"Classe: Kina",
		"Mana total: 4.770.000",
		"Potions: 5.300 a 7.950",
		"Custo: 3.510.000 a 5.200.000 gold",
	}
	for _, part := range expectedParts {
		if !strings.Contains(text, part) {
			t.Errorf("expected upskill mana reply to contain %q, got %q", part, text)
		}
	}
	if strings.Contains(strings.ToLower(text), "pack") {
		t.Errorf("expected upskill mana reply not to mention packs, got %q", text)
	}
	if strings.Contains(text, "Treinando 8h por dia") {
		t.Errorf("expected no daily training estimate when daily hours are omitted, got %q", text)
	}
	if sentImageCount(fake) != 1 {
		t.Errorf("expected one generated upskill card image, got %d", sentImageCount(fake))
	}
	if sentTextCount(fake) != 0 {
		t.Errorf("expected upskill text to be sent as the image caption, got %d text messages", sentTextCount(fake))
	}
}

func TestUpskillWithManaEstimateAndDailyHoursInAnyOrder(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"vocation then daily hours", []string{"400", "450", "5000", "kina", "8"}},
		{"daily hours then vocation", []string{"400", "450", "5000", "8", "kina"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withRucoyServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("01:02:30"))
			})

			fake := &botfakes.FakeClient{}
			m := newTestMessage(t, fake, newTestDB(t))
			*m.Args = tt.args

			Upskill(m)

			text := lastReplyText(fake)
			if !strings.Contains(text, "Treinando 8h por dia: 3 dias, 2 horas e 30 minutos") ||
				!strings.Contains(text, "Classe: Kina") ||
				!strings.Contains(text, "Custo: 3.510.000 a 5.200.000 gold") {
				t.Errorf("expected daily and mana estimates for args %v, got %q", tt.args, text)
			}
			if sentImageCount(fake) != 1 {
				t.Errorf("expected one generated upskill card image for args %v, got %d", tt.args, sentImageCount(fake))
			}
		})
	}
}

func TestUpskillWithPallyAddsArrowCostInAnyOrder(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"vocation then daily hours", []string{"400", "450", "5000", "pally", "8"}},
		{"daily hours then vocation", []string{"400", "450", "5000", "8", "pally"}},
		{"vocation only", []string{"400", "450", "5000", "pally"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withRucoyServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("01:02:30"))
			})

			fake := &botfakes.FakeClient{}
			m := newTestMessage(t, fake, newTestDB(t))
			*m.Args = tt.args

			Upskill(m)

			text := lastReplyText(fake)
			expectedParts := []string{
				"Classe: Pally",
				"Flechas: 381.600",
				"Custo flechas: 764.000 gold",
				"Custo total: 4.274.000 a 5.964.000 gold",
			}
			for _, part := range expectedParts {
				if !strings.Contains(text, part) {
					t.Errorf("expected pally upskill reply to contain %q for args %v, got %q", part, tt.args, text)
				}
			}
			if sentImageCount(fake) != 1 {
				t.Errorf("expected one generated pally upskill card image for args %v, got %d", tt.args, sentImageCount(fake))
			}
		})
	}
}

func TestUpskillInvalidDailyHours(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"400", "450", "5000", "25"}

	Upskill(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "Opcional invalido") {
		t.Errorf("expected daily hours validation error, got %q", text)
	}
}

func TestUpskillHTTPError(t *testing.T) {
	withRucoyServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"400", "450", "5000"}

	Upskill(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "Erro ao calcular") {
		t.Errorf("expected an error reply, got %q", text)
	}
}

func TestUplevelSuccess(t *testing.T) {
	withRucoyServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("1000000"))
	})

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"350", "400", "20kk"}

	Uplevel(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "Uplevel Rucoy") {
		t.Errorf("expected a formatted uplevel reply, got %q", text)
	}
}

func TestUplevelWithDailyHours(t *testing.T) {
	withRucoyServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("530000000"))
	})

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"350", "400", "20kk", "8"}

	Uplevel(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "Tempo estimado: 26 horas e 30 minutos") ||
		!strings.Contains(text, "Treinando 8h por dia: 3 dias, 2 horas e 30 minutos") {
		t.Errorf("expected daily training estimate in uplevel reply, got %q", text)
	}
}

func TestUplevelInvalidDailyHours(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"350", "400", "20kk", "0"}

	Uplevel(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "horas_por_dia") {
		t.Errorf("expected daily hours validation error, got %q", text)
	}
}

func TestUplevelInvalidLevels(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"400", "350", "20kk"}

	Uplevel(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "maior que o level atual") {
		t.Errorf("expected a validation error reply, got %q", text)
	}
}
