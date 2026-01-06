package utils

import (
	"log/slog"
	"os"
	"strings"
)

func ReadDotEnv() {
	file, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	slog.Warn(".env file found")
	for line := range strings.SplitSeq(string(file), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		items := strings.Split(line, "=")
		items[1] = strings.Join(items[1:], "=")
		items = items[:2]
		if len(items) != 2 {
			panic("Inconsistent envs found")
		}
		os.Setenv(items[0], items[1])
	}
}
