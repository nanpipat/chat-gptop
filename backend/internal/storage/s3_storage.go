package storage

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Storage implements Storage for AWS S3 compatible object storage (like Cloudflare R2)
type S3Storage struct {
	client *s3.Client
	bucket string
}

// NewS3Storage creates a new S3Storage instance
// It configures the client to work with Cloudflare R2 or other S3-compatible providers
func NewS3Storage(accessKey, secretKey, endpoint, bucket string) (*S3Storage, error) {
	// Force IPv4 transport to resolve connectivity issues in Docker environments
	// similar to what we did for Supabase connection (PostgreSQL)
	customTransport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext(ctx, "tcp4", addr)
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// Force HTTP/1.1 (Disable HTTP/2) as R2 seems to have TLS handshake issues with HTTP/2 in some environments
		ForceAttemptHTTP2: false,
		// This empty map disables HTTP/2 support in Transport
		TLSNextProto: make(map[string]func(string, *tls.Conn) http.RoundTripper),
	}

	customClient := &http.Client{
		Transport: customTransport,
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion("auto"), // R2 uses 'auto' region
		config.WithHTTPClient(customClient),
	)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // Required for R2 and some S3-compatible services
	})

	return &S3Storage{
		client: client,
		bucket: bucket,
	}, nil
}

func (s *S3Storage) Put(ctx context.Context, path string, content io.Reader) error {
	// Clean path to remove leading/trailing slashes if any, R2 treats keys as is
	key := filepath.Clean(path)
	if key == "." || key == "/" {
		return fmt.Errorf("invalid path: %s", path)
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   content,
	})
	if err != nil {
		log.Printf("[S3Storage] PutObject failed: bucket=%s key=%s error=%v", s.bucket, key, err)
	}
	return err
}

func (s *S3Storage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	key := filepath.Clean(path)

	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *S3Storage) Delete(ctx context.Context, path string) error {
	key := filepath.Clean(path)

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("[S3Storage] DeleteObject failed: bucket=%s key=%s error=%v", s.bucket, key, err)
	}
	return err
}
