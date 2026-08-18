package filter

import "testing"

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		pattern string
		want    bool
	}{
		{"exact match", "antedeguemon", "antedeguemon", true},
		{"match as whole word within sentence", "ola antedeguemon tudo bem", "antedeguemon", true},
		{"case insensitive", "ANTEDEGUEMON", "antedeguemon", true},
		{"substring is not a match", "antedeguemonzão", "antedeguemon", false},
		{"no match", "algo completamente diferente", "antedeguemon", false},
		{"pattern at start of text", "antedeguemon foi ativado", "antedeguemon", true},
		{"pattern at end of text", "foi ativado antedeguemon", "antedeguemon", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesPattern(tt.text, tt.pattern); got != tt.want {
				t.Errorf("matchesPattern(%q, %q) = %v, want %v", tt.text, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestGetCompiledPatternIsCached(t *testing.T) {
	pattern := "uniquetestpattern123"
	defer FilterCache.Delete(pattern)

	re1, err := getCompiledPattern(pattern)
	if err != nil {
		t.Fatalf("getCompiledPattern() error = %v", err)
	}
	re2, err := getCompiledPattern(pattern)
	if err != nil {
		t.Fatalf("getCompiledPattern() error = %v", err)
	}

	if re1 != re2 {
		t.Error("expected the same compiled regexp instance to be returned from the cache")
	}

	if _, ok := FilterCache.Load(pattern); !ok {
		t.Error("expected pattern to be present in FilterCache after compilation")
	}
}

func TestGetCompiledPatternEscapesRegexMetacharacters(t *testing.T) {
	pattern := "a.b*c"
	defer FilterCache.Delete(pattern)

	re, err := getCompiledPattern(pattern)
	if err != nil {
		t.Fatalf("getCompiledPattern() error = %v", err)
	}

	if !re.MatchString("a.b*c") {
		t.Error("expected literal pattern to match itself")
	}
	if re.MatchString("aXbYYYc") {
		t.Error("regex metacharacters in the pattern should be escaped, not interpreted")
	}
}
