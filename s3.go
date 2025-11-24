package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// loadAWSConfig returns an AWS config
func loadAWSConfig() (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(getRegion()),
	)
	if err != nil {
		return aws.Config{}, err
	}
	return cfg, nil
}

// getRegion returns the AWS region from environment or default
func getRegion() string {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		// Default fallback if not set, but prefer env var
		return "us-east-1"
	}
	return region
}

// UploadSSEKMS uploads `body` to s3 at key `key` using SSE-KMS with kmsKeyID.
func UploadSSEKMS(cfg aws.Config, bucket, key string, body []byte, kmsKeyID string) error {
	s3c := s3.NewFromConfig(cfg)
	input := &s3.PutObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(key),
		Body:                 bytes.NewReader(body),
		ServerSideEncryption: s3types.ServerSideEncryptionAwsKms,
		SSEKMSKeyId:          aws.String(kmsKeyID),
	}
	_, err := s3c.PutObject(context.Background(), input)
	return err
}

// UploadRaw uploads body to S3 at key with optional metadata
func UploadRaw(cfg aws.Config, bucket, key string, body []byte, metadata map[string]string) error {
	s3c := s3.NewFromConfig(cfg)
	input := &s3.PutObjectInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		Body:     bytes.NewReader(body),
		Metadata: metadata,
	}
	_, err := s3c.PutObject(context.Background(), input)
	return err
}

// ListKeys lists up to `max` keys under the prefix
func ListKeys(cfg aws.Config, bucket, prefix string, max int32) ([]string, error) {
	s3c := s3.NewFromConfig(cfg)
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(max),
	}
	out, err := s3c.ListObjectsV2(context.Background(), input)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(out.Contents))
	for _, o := range out.Contents {
		if o.Key != nil {
			keys = append(keys, *o.Key)
		}
	}
	return keys, nil
}

// GetObjectBytes fetches object content (all bytes)
func GetObjectBytes(cfg aws.Config, bucket, key string) ([]byte, map[string]string, error) {
	s3c := s3.NewFromConfig(cfg)
	out, err := s3c.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, nil, err
	}
	defer out.Body.Close()
	b, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, nil, err
	}
	meta := map[string]string{}
	if out.Metadata != nil {
		for k, v := range out.Metadata {
			meta[k] = v
		}
	}
	return b, meta, nil
}

// S3URL Helper: s3 URL
func S3URL(bucket, key string) string {
	// return s3://bucket/key
	return fmt.Sprintf("s3://%s/%s", bucket, url.PathEscape(key))
}
