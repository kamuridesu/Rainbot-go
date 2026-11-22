package configs

import (
	"cmp"
	"errors"
	"os"
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
	ErrMissingDB     = errors.New("DB_DRIVER and DB_PARAMS environment variables cannot be empty")
	ErrMissingBucket = errors.New("S3_BUCKET_NAME must be set when STORAGE_TYPE is S3")
)

func env(key, def string) string {
	return cmp.Or(os.Getenv(key), def)
}

func Load() *Config {
	return &Config{
		Bot: Bot{
			Name:              env("BOT_NAME", "teto"),
			OwnerNumber:       os.Getenv("CREATOR_NUMBER"),
			BroadcastPassword: os.Getenv("BROADCAST_PASSWORD"),
			QuotlyUrl:         os.Getenv("QUOTLY_API_URL"),
		},
		DB: DB{
			Driver: os.Getenv("DB_DRIVER"),
			Params: os.Getenv("DB_PARAMS"),
		},
		Storage: Storage{
			Type: StorageType(env("STORAGE_TYPE", string(StorageLocal))),
			Aws: AWS{
				Region:          env("AWS_REGION", "us-west-2"),
				EndpointUrl:     os.Getenv("AWS_ENDPOINT_URL"),
				BucketName:      os.Getenv("S3_BUCKET_NAME"),
				AccessKeyId:     os.Getenv("AWS_ACCESS_KEY_ID"),
				SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
			},
		},
		Cobalt: Cobalt{
			Url:    os.Getenv("COBALT_URL"),
			ApiKey: os.Getenv("COBALT_API_KEY"),
		},
		AI: AI{
			ModelKey: os.Getenv("AI_MODEL_KEY"),
			ModelUrl: os.Getenv("AI_MODEL_URL"),
		},
	}
}

func (c *Config) Validate() error {
	var errs []error
	if c.DB.Driver == "" || c.DB.Params == "" {
		errs = append(errs, ErrMissingDB)
	}
	if c.Storage.Type == StorageS3 && c.Storage.Aws.BucketName == "" {
		errs = append(errs, ErrMissingBucket)
	}
	return errors.Join(errs...)
}

var Get = sync.OnceValue(Load)
