package fun

import (
	"net/http"
	"testing"
	"time"
)

func TestParseRucoyResponse(t *testing.T) {
	data := `<table>
<tr><td><a href="/characters/John">John</a></td><td>50</td><td><span class="status">Online</span></td></tr>
<tr><td><a href="/characters/Jane">Jane</a></td><td>30</td><td><span class="status">Offline</span></td></tr>
</table>`

	got := parseRucoyResponse(data, "TestGuild")

	if got.Guild != "TestGuild" {
		t.Errorf("Guild = %q, want %q", got.Guild, "TestGuild")
	}
	if len(got.Members) != 2 {
		t.Fatalf("expected 2 members, got %d: %+v", len(got.Members), got.Members)
	}

	john := got.Members[0]
	if john.Name != "John" || john.Level != 50 || !john.Online || john.CharacterPath != "/characters/John" {
		t.Errorf("unexpected first member: %+v", john)
	}

	jane := got.Members[1]
	if jane.Name != "Jane" || jane.Level != 30 || jane.Online || jane.CharacterPath != "/characters/Jane" {
		t.Errorf("unexpected second member: %+v", jane)
	}
}

func TestParseRucoyResponseNoMatches(t *testing.T) {
	got := parseRucoyResponse("<html><body>no rows here</body></html>", "EmptyGuild")
	if len(got.Members) != 0 {
		t.Errorf("expected no members, got %d", len(got.Members))
	}
}

func TestStripRucoyHTML(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"strips tags and collapses whitespace", "<b>Hello</b>  world", "Hello world"},
		{"unescapes html entities", "A &amp; B", "A & B"},
		{"nested tags", "<span><b>Bold</b> text</span>", "Bold text"},
		{"plain text unchanged", "already clean", "already clean"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripRucoyHTML(tt.value); got != tt.want {
				t.Errorf("stripRucoyHTML(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestParseRucoyCharacterInfo(t *testing.T) {
	data := `<table>
<tr><td>Name</td><td>John</td></tr>
<tr><td>Level</td><td>123</td></tr>
<tr><td>Guild</td><td>TestGuild</td></tr>
<tr><td>Title</td><td>Hero</td></tr>
<tr><td>Last Online</td><td>Online</td></tr>
</table>`

	got := parseRucoyCharacterInfo(data)

	if got.Name != "John" {
		t.Errorf("Name = %q, want %q", got.Name, "John")
	}
	if got.Level != 123 {
		t.Errorf("Level = %d, want %d", got.Level, 123)
	}
	if got.Guild != "TestGuild" {
		t.Errorf("Guild = %q, want %q", got.Guild, "TestGuild")
	}
	if got.Title != "Hero" {
		t.Errorf("Title = %q, want %q", got.Title, "Hero")
	}
	if got.LastOnline != "Online" {
		t.Errorf("LastOnline = %q, want %q", got.LastOnline, "Online")
	}
}

func TestParseRucoyLastOnlineDays(t *testing.T) {
	tests := []struct {
		name string
		data string
		want int
	}{
		{"currently online", `<td>Last online</td><td>Online</td>`, 0},
		{"minutes ago rounds to 0", `<td>Last online</td><td>10 minutes ago</td>`, 0},
		{"hours ago rounds to 0", `<td>Last online</td><td>3 hours ago</td>`, 0},
		{"days ago", `<td>Last online</td><td>5 days ago</td>`, 5},
		{"weeks ago converted to days", `<td>Last online</td><td>2 weeks ago</td>`, 14},
		{"months ago converted to days", `<td>Last online</td><td>3 months ago</td>`, 90},
		{"years ago converted to days", `<td>Last online</td><td>1 year ago</td>`, 365},
		{"missing row", `<td>Something else</td><td>5 days ago</td>`, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseRucoyLastOnlineDays(tt.data); got != tt.want {
				t.Errorf("parseRucoyLastOnlineDays(%q) = %d, want %d", tt.data, got, tt.want)
			}
		})
	}
}

func TestFormatUpskillTime(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"days:hours:minutes", "1:02:03", "26 horas e 3 minutos"},
		{"hours:minutes", "05:30", "5 horas e 30 minutos"},
		{"minutes only", "45", "45 minutos"},
		{"unparseable falls back to raw", "1:2:3:4", "1:2:3:4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatUpskillTime(tt.raw); got != tt.want {
				t.Errorf("formatUpskillTime(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestParseRucoyDailyHours(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    int
		wantErr bool
	}{
		{"empty optional argument", nil, 0, false},
		{"valid daily hours", []string{"8"}, 8, false},
		{"zero is invalid", []string{"0"}, 0, true},
		{"above 24 is invalid", []string{"25"}, 0, true},
		{"text is invalid", []string{"abc"}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRucoyDailyHours(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseRucoyDailyHours(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("parseRucoyDailyHours(%v) = %d, want %d", tt.args, got, tt.want)
			}
		})
	}
}

func TestFormatRucoyDailyTrainingEstimate(t *testing.T) {
	tests := []struct {
		name          string
		estimatedTime string
		dailyHours    int
		want          string
	}{
		{"exact days", "24 horas e 0 minutos", 8, "Treinando 8h por dia: 3 dias"},
		{"days with remainder", "26 horas e 30 minutos", 8, "Treinando 8h por dia: 3 dias, 2 horas e 30 minutos"},
		{"less than one daily session", "5 horas e 30 minutos", 8, "Treinando 8h por dia: 5 horas e 30 minutos"},
		{"minutes only", "45 minutos", 8, "Treinando 8h por dia: 45 minutos"},
		{"unparseable returns empty", "???", 8, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatRucoyDailyTrainingEstimate(tt.estimatedTime, tt.dailyHours); got != tt.want {
				t.Errorf("formatRucoyDailyTrainingEstimate(%q, %d) = %q, want %q", tt.estimatedTime, tt.dailyHours, got, tt.want)
			}
		})
	}
}

func TestValidateRucoyTrainInput(t *testing.T) {
	tests := []struct {
		name             string
		attack           int
		baseLevel        int
		statLevel        int
		extraStats       int
		targetEfficiency float64
		wantErr          bool
	}{
		{"valid input from command example", 5, 351, 391, -50, 90, false},
		{"baseLevel too low", 5, 0, 391, -50, 90, true},
		{"baseLevel too high", 5, 1001, 391, -50, 90, true},
		{"statLevel too low", 5, 351, 4, -50, 90, true},
		{"extraStats too low", 5, 351, 391, -81, 90, true},
		{"extraStats too high", 5, 351, 391, 127, 90, true},
		{"attack too low", 3, 351, 391, -50, 90, true},
		{"attack too high", 61, 351, 391, -50, 90, true},
		{"attack on excluded tier (6)", 6, 351, 391, -50, 90, true},
		{"attack boundary valid (4)", 4, 351, 391, -50, 90, false},
		{"attack boundary valid (60)", 60, 351, 391, -50, 90, false},
		{"efficiency too low", 5, 351, 391, -50, 34, true},
		{"efficiency too high", 5, 351, 391, -50, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRucoyTrainInput(tt.attack, tt.baseLevel, tt.statLevel, tt.extraStats, tt.targetEfficiency)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRucoyTrainInput(%d,%d,%d,%d,%v) error = %v, wantErr %v",
					tt.attack, tt.baseLevel, tt.statLevel, tt.extraStats, tt.targetEfficiency, err, tt.wantErr)
			}
		})
	}
}

func TestCalculateRucoyTrainingIsDeterministic(t *testing.T) {
	result1 := calculateRucoyTraining(351, 391, -50, 5, false, 90)
	result2 := calculateRucoyTraining(351, 391, -50, 5, false, 90)

	if result1 != result2 {
		t.Errorf("calculateRucoyTraining is not deterministic for identical inputs:\n%+v\n%+v", result1, result2)
	}

	if result1.Mode != "AFK Train" {
		t.Errorf("Mode = %q, want %q", result1.Mode, "AFK Train")
	}
	if result1.MinimumDuration != rucoyMinimumTrainingDurationSeconds {
		t.Errorf("MinimumDuration = %v, want %v", result1.MinimumDuration, rucoyMinimumTrainingDurationSeconds)
	}
}

func TestCalculateRucoyTrainingPowertrainMode(t *testing.T) {
	result := calculateRucoyTraining(351, 391, -50, 5, true, 90)
	if result.Mode != "Powertrain" {
		t.Errorf("Mode = %q, want %q", result.Mode, "Powertrain")
	}
}

func TestCalculateRucoyTrainingHighEfficiencyMayFindNoMonster(t *testing.T) {
	result := calculateRucoyTraining(1, 5, -80, 4, false, 99)
	if result.DurationSeconds < 0 {
		t.Errorf("DurationSeconds should never be negative, got %v", result.DurationSeconds)
	}
}

func TestFormatRucoyTrainingDuration(t *testing.T) {
	tests := []struct {
		seconds float64
		want    string
	}{
		{0, "00:00"},
		{65, "01:05"},
		{600, "10:00"},
		{3661, "61:01"},
	}

	for _, tt := range tests {
		if got := formatRucoyTrainingDuration(tt.seconds); got != tt.want {
			t.Errorf("formatRucoyTrainingDuration(%v) = %q, want %q", tt.seconds, got, tt.want)
		}
	}
}

func TestParseRucoyXPPerHour(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    int64
		wantErr bool
	}{
		{"plain integer", "5000", 5000, false},
		{"kk suffix means million", "20kk", 20000000, false},
		{"m suffix means million", "20m", 20000000, false},
		{"k suffix means thousand", "5k", 5000, false},
		{"comma as decimal separator", "1,5k", 1500, false},
		{"whitespace and underscores stripped", "5_000 ", 5000, false},
		{"uppercase suffix", "5K", 5000, false},
		{"not a number", "abc", 0, true},
		{"zero is invalid", "0", 0, true},
		{"negative is invalid", "-5", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRucoyXPPerHour(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseRucoyXPPerHour(%q) error = %v, wantErr %v", tt.raw, err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("parseRucoyXPPerHour(%q) = %d, want %d", tt.raw, got, tt.want)
			}
		})
	}
}

func TestFormatRucoyDuration(t *testing.T) {
	tests := []struct {
		name      string
		xpNeeded  int64
		xpPerHour int64
		want      string
	}{
		{"less than an hour", 1000, 4000, "15 minutos"},
		{"exact hour", 4000, 4000, "1 horas e 0 minutos"},
		{"multiple hours", 10000, 4000, "2 horas e 30 minutos"},
		{"rounds up partial minute", 1, 4000, "1 minutos"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatRucoyDuration(tt.xpNeeded, tt.xpPerHour); got != tt.want {
				t.Errorf("formatRucoyDuration(%d, %d) = %q, want %q", tt.xpNeeded, tt.xpPerHour, got, tt.want)
			}
		})
	}
}

func TestAbsRucoyNumber(t *testing.T) {
	tests := []struct {
		value int64
		want  int64
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
	}
	for _, tt := range tests {
		if got := absRucoyNumber(tt.value); got != tt.want {
			t.Errorf("absRucoyNumber(%d) = %d, want %d", tt.value, got, tt.want)
		}
	}
}

func TestFormatRucoyNumber(t *testing.T) {
	tests := []struct {
		value int64
		want  string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1.000"},
		{1234567, "1.234.567"},
	}
	for _, tt := range tests {
		if got := formatRucoyNumber(tt.value); got != tt.want {
			t.Errorf("formatRucoyNumber(%d) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestRucoyRetryDelay(t *testing.T) {
	fallback := 30 * time.Second

	t.Run("no header returns fallback", func(t *testing.T) {
		res := &http.Response{Header: http.Header{}}
		if got := rucoyRetryDelay(res, fallback); got != fallback {
			t.Errorf("rucoyRetryDelay() = %v, want fallback %v", got, fallback)
		}
	})

	t.Run("numeric seconds header", func(t *testing.T) {
		res := &http.Response{Header: http.Header{"Retry-After": []string{"5"}}}
		if got := rucoyRetryDelay(res, fallback); got != 5*time.Second {
			t.Errorf("rucoyRetryDelay() = %v, want %v", got, 5*time.Second)
		}
	})

	t.Run("invalid header returns fallback", func(t *testing.T) {
		res := &http.Response{Header: http.Header{"Retry-After": []string{"not-a-value"}}}
		if got := rucoyRetryDelay(res, fallback); got != fallback {
			t.Errorf("rucoyRetryDelay() = %v, want fallback %v", got, fallback)
		}
	})

	t.Run("http-date in the past returns fallback", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Hour).UTC().Format(http.TimeFormat)
		res := &http.Response{Header: http.Header{"Retry-After": []string{past}}}
		if got := rucoyRetryDelay(res, fallback); got != fallback {
			t.Errorf("rucoyRetryDelay() = %v, want fallback %v", got, fallback)
		}
	})

	t.Run("http-date in the future returns remaining duration", func(t *testing.T) {
		future := time.Now().Add(10 * time.Second).UTC().Format(http.TimeFormat)
		res := &http.Response{Header: http.Header{"Retry-After": []string{future}}}
		got := rucoyRetryDelay(res, fallback)
		if got <= 0 || got > 11*time.Second {
			t.Errorf("rucoyRetryDelay() = %v, want roughly 10s", got)
		}
	})
}
