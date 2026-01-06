package storage

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3FileStorage struct {
	svc    *s3.Client
	bucket string
}

var _s3Singleton *S3FileStorage

func NewS3FileStorage() FileStorage {
	bucket := os.Getenv("S3_BUCKET_NAME")
	if bucket == "" {
		panic("Environment variable S3_BUCKET_NAME not set")
	}
	if _s3Singleton != nil {
		return _s3Singleton
	}

	region := cmp.Or(os.Getenv("AWS_REGION"), "us-west-2")

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		panic(err)
	}
	svc := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		}
	})
	_s3Singleton = &S3FileStorage{svc: svc, bucket: bucket}
	return _s3Singleton
}

func (s *S3FileStorage) Write(ctx context.Context, filename string, data []byte) error {
	input := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(filename),
		Body:          bytes.NewReader(data),
		ContentLength: aws.Int64(int64(len(data))),
	}

	_, err := s.svc.PutObject(ctx, input)
	return err
}

func (s *S3FileStorage) Delete(ctx context.Context, filename string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
	}
	_, err := s.svc.DeleteObject(ctx, input)
	return err
}

func (s *S3FileStorage) Exists(ctx context.Context, filename string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
	}
	_, err := s.svc.HeadObject(ctx, input)
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *S3FileStorage) Read(ctx context.Context, filename string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
	}
	data, err := s.svc.GetObject(ctx, input)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(data.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
