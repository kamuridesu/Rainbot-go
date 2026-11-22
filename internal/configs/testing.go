package configs

import "testing"

func Override(t testing.TB, fn func(*Config)) {
	t.Helper()
	cfg := Get()
	orig := *cfg
	fn(cfg)
	t.Cleanup(func() { *cfg = orig })
}
