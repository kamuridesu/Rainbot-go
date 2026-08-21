package rucoy

import (
	"strings"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/commands"
)

func TestRucoyCommandsHaveTutorialHelp(t *testing.T) {
	tests := []struct {
		command       string
		description   []string
		exampleParts  []string
		expectedAlias string
	}{
		{
			command:       "online",
			description:   []string{"Uso:", "/online Nome-da-Guild"},
			exampleParts:  []string{"${prefix}${alias} Nome-da-Guild", "${prefix}${alias} B L A C K O U T"},
			expectedAlias: "ronline",
		},
		{
			command:      "upskill",
			description:  []string{"Uso:", "/upskill skill_atual skill_desejada tickrate [classe] [horas_por_dia]", "Classes:", "pally = 50 mana por skill + flechas"},
			exampleParts: []string{"${prefix}${alias} 400 450 42000", "${prefix}${alias} 400 450 42000 kina", "${prefix}${alias} 400 450 42000 pally 8", "${prefix}${alias} 400 450 42000 8 mage"},
		},
		{
			command:      "uplevel",
			description:  []string{"Uso:", "/uplevel level_atual level_desejado xp_por_hora [horas_por_dia]"},
			exampleParts: []string{"${prefix}${alias} 350 400 20kk", "${prefix}${alias} 350 400 20kk 8", "${prefix}${alias} 275 300 5kk"},
		},
		{
			command:      "train",
			description:  []string{"Uso:", "/train arma level skill add [eficiencia]", "Armas de treino comuns:"},
			exampleParts: []string{"${prefix}${alias} 5 351 391 -50", "${prefix}${alias} 5 351 391 -50 90", "${prefix}${alias} 7 400 450 0"},
		},
		{
			command:      "afk",
			description:  []string{"Uso:", "/afk Nome-da-Guild"},
			exampleParts: []string{"${prefix}${alias} Nome-da-Guild", "${prefix}${alias} B L A C K O U T"},
		},
		{
			command:      "info",
			description:  []string{"Uso:", "/info Nome-do-Jogador", "Black Skull", "XP Mobwin"},
			exampleParts: []string{"${prefix}${alias} Nome-do-Jogador", "${prefix}${alias} Kamuri SG"},
		},
		{
			command:      "meta",
			description:  []string{"Uso:", "/meta level_meta Nome-da-Guild"},
			exampleParts: []string{"${prefix}${alias} 400 Nome-da-Guild", "${prefix}${alias} 500 B L A C K O U T"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			cmd, err := commands.FindCommand(tt.command)
			if err != nil {
				t.Fatalf("FindCommand(%q) error = %v", tt.command, err)
			}

			for _, part := range tt.description {
				if !strings.Contains(cmd.Description, part) {
					t.Errorf("expected %q help description to contain %q, got %q", tt.command, part, cmd.Description)
				}
			}

			for _, example := range tt.exampleParts {
				if !containsString(*cmd.Examples, example) {
					t.Errorf("expected %q examples to contain %q, got %v", tt.command, example, *cmd.Examples)
				}
			}

			if tt.expectedAlias != "" && !containsString(*cmd.Aliases, tt.expectedAlias) {
				t.Errorf("expected %q aliases to contain %q, got %v", tt.command, tt.expectedAlias, *cmd.Aliases)
			}
		})
	}
}

func TestRucoyCategoryHasMenuBanner(t *testing.T) {
	if RucoyCategory.BannerPath == nil {
		t.Fatal("expected RucoyCategory to have a menu banner")
	}
	if *RucoyCategory.BannerPath != "assets/rucoy/menu-banner.png" {
		t.Fatalf("BannerPath = %q, want %q", *RucoyCategory.BannerPath, "assets/rucoy/menu-banner.png")
	}

	t.Chdir(repoRoot(t))
	banner, err := RucoyCategory.LoadBanner()
	if err != nil {
		t.Fatalf("LoadBanner() error = %v", err)
	}
	if len(banner) == 0 {
		t.Fatal("expected Rucoy menu banner to have content")
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
