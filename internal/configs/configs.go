package configs

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

type StorageType string

const (
	StorageLocal StorageType = "LOCAL"
	StorageS3    StorageType = "S3"
)

type Bot struct {
	Name              string
	OwnerNumber       string
	BroadcastPassword string
	QuotlyUrl         string
}

type DB struct {
	Driver string
	Params string
}

type AWS struct {
	Region          string
	EndpointUrl     string
	BucketName      string
	AccessKeyId     string
	SecretAccessKey string
}

type Storage struct {
	Type StorageType
	Aws  AWS
}

type Cobalt struct {
	Url    string
	ApiKey string
}

type AI struct {
	ModelKey string
	ModelUrl string
}

type Config struct {
	Bot     Bot
	DB      DB
	Storage Storage
	Cobalt  Cobalt
	AI      AI
}

var (
	ErrMissingDB          = errors.New("DB_DRIVER and DB_PARAMS environment variables cannot be empty")
	ErrUnsupportedDB      = errors.New("DB_DRIVER must be one of: sqlite3, postgres")
	ErrMissingBucket      = errors.New("S3_BUCKET_NAME must be set when STORAGE_TYPE is S3")
	ErrInvalidStorageType = errors.New("STORAGE_TYPE must be one of: LOCAL, S3")
	ErrPartialAWSKeys     = errors.New("AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY must be set together")
)

func env(key, def string) string {
	return cmp.Or(strings.TrimSpace(os.Getenv(key)), def)
}

func Load() *Config {
	return &Config{
		Bot: Bot{
			Name:              env("BOT_NAME", "teto"),
			OwnerNumber:       env("CREATOR_NUMBER", ""),
			BroadcastPassword: env("BROADCAST_PASSWORD", ""),
			QuotlyUrl:         env("QUOTLY_API_URL", ""),
		},
		DB: DB{
			Driver: env("DB_DRIVER", ""),
			Params: env("DB_PARAMS", ""),
		},
		Storage: Storage{
			Type: StorageType(env("STORAGE_TYPE", string(StorageLocal))),
			Aws: AWS{
				Region:          env("AWS_REGION", "us-west-2"),
				EndpointUrl:     env("AWS_ENDPOINT_URL", ""),
				BucketName:      env("S3_BUCKET_NAME", ""),
				AccessKeyId:     env("AWS_ACCESS_KEY_ID", ""),
				SecretAccessKey: env("AWS_SECRET_ACCESS_KEY", ""),
			},
		},
		Cobalt: Cobalt{
			Url:    env("COBALT_URL", ""),
			ApiKey: env("COBALT_API_KEY", ""),
		},
		AI: AI{
			ModelKey: env("AI_MODEL_KEY", ""),
			ModelUrl: env("AI_MODEL_URL", ""),
		},
	}
}

func (c *Config) Validate() error {
	var errs []error
	if c.DB.Driver == "" || c.DB.Params == "" {
		errs = append(errs, ErrMissingDB)
	} else if c.DB.Driver != "sqlite3" && c.DB.Driver != "postgres" {
		errs = append(errs, fmt.Errorf("%w (got %q)", ErrUnsupportedDB, c.DB.Driver))
	}

	switch c.Storage.Type {
	case StorageLocal:
	case StorageS3:
		if c.Storage.Aws.BucketName == "" {
			errs = append(errs, ErrMissingBucket)
		}
		if (c.Storage.Aws.AccessKeyId == "") != (c.Storage.Aws.SecretAccessKey == "") {
			errs = append(errs, ErrPartialAWSKeys)
		}
	default:
		errs = append(errs, fmt.Errorf("%w (got %q)", ErrInvalidStorageType, c.Storage.Type))
	}
	return errors.Join(errs...)
}

func (c *Config) Warnings() []string {
	var w []string
	if c.Bot.OwnerNumber == "" {
		w = append(w, "CREATOR_NUMBER not set: bug reports cannot be delivered")
	}
	if c.Bot.BroadcastPassword == "" {
		w = append(w, "BROADCAST_PASSWORD not set: broadcast command is disabled")
	}
	if c.Bot.QuotlyUrl == "" {
		w = append(w, "QUOTLY_API_URL not set: quote stickers are disabled")
	}
	if c.Cobalt.Url == "" || c.Cobalt.ApiKey == "" {
		w = append(w, "COBALT_URL or COBALT_API_KEY not set: media downloads are disabled")
	}
	if c.AI.ModelKey == "" {
		w = append(w, "AI_MODEL_KEY not set: AI chat replies are disabled")
	}
	return w
}

func (c *Config) AIEnabled() bool {
	return c.AI.ModelKey != ""
}

var Get = sync.OnceValue(Load)
