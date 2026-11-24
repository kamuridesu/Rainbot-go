package configs

import (
	"errors"
	"testing"
)

func validConfig() *Config {
	return &Config{
		DB:      DB{Driver: "sqlite3", Params: "file:test.db"},
		Storage: Storage{Type: StorageLocal},
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
		want   error
	}{
		{"valid", func(*Config) {}, nil},
		{"missing db params", func(c *Config) { c.DB.Params = "" }, ErrMissingDB},
		{"unsupported driver", func(c *Config) { c.DB.Driver = "mysql" }, ErrUnsupportedDB},
		{"invalid storage type", func(c *Config) { c.Storage.Type = "GCS" }, ErrInvalidStorageType},
		{"s3 without bucket", func(c *Config) { c.Storage.Type = StorageS3 }, ErrMissingBucket},
		{"s3 partial keys", func(c *Config) {
			c.Storage.Type = StorageS3
			c.Storage.Aws.BucketName = "b"
			c.Storage.Aws.AccessKeyId = "id"
		}, ErrPartialAWSKeys},
		{"s3 valid with default credentials", func(c *Config) {
			c.Storage.Type = StorageS3
			c.Storage.Aws.BucketName = "b"
		}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig()
			tt.mutate(c)
			err := c.Validate()
			if tt.want == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
		})
	}
}

func TestEnvTrimsBlankValues(t *testing.T) {
	t.Setenv("RAINBOT_TEST_VAR", "   ")
	if got := env("RAINBOT_TEST_VAR", "def"); got != "def" {
		t.Fatalf("got %q, want %q", got, "def")
	}
	t.Setenv("RAINBOT_TEST_VAR", "  value ")
	if got := env("RAINBOT_TEST_VAR", "def"); got != "value" {
		t.Fatalf("got %q, want %q", got, "value")
	}
}

func TestWarnings(t *testing.T) {
	c := validConfig()
	if got := len(c.Warnings()); got != 5 {
		t.Fatalf("got %d warnings, want 5", got)
	}
	c.Bot = Bot{OwnerNumber: "1", BroadcastPassword: "p", QuotlyUrl: "u"}
	c.Cobalt = Cobalt{Url: "u", ApiKey: "k"}
	c.AI = AI{ModelKey: "k"}
	if w := c.Warnings(); len(w) != 0 {
		t.Fatalf("expected no warnings, got %v", w)
	}
}
