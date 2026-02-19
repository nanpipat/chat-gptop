package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	// Load .env from parent directory
	loadEnv("../../.env")

	endpoint := os.Getenv("R2_ENDPOINT")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucket := os.Getenv("R2_BUCKET")

	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		fmt.Println("Error: Missing R2 environment variables in .env")
		return
	}

	fmt.Printf("Testing R2 Connection...\n")
	fmt.Printf("Endpoint: %s\n", endpoint)
	fmt.Printf("Bucket: %s\n", bucket)
	fmt.Println("---------------------------------------------------")

	// Test 1: Standard
	fmt.Println("\n[Test 1] Standard Connection (Default Transport)")
	testConnection(endpoint, accessKey, secretKey, bucket, nil)

	// Test 2: Force IPv4
	fmt.Println("\n[Test 2] Force IPv4 (tcp4)")
	ipv4Transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext(ctx, "tcp4", addr)
		},
	}
	testConnection(endpoint, accessKey, secretKey, bucket, ipv4Transport)

	// Test 3: Insecure TLS (Skip Verify)
	fmt.Println("\n[Test 3] Insecure TLS (Skip Verify)")
	insecureTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	testConnection(endpoint, accessKey, secretKey, bucket, insecureTransport)

	// Test 4: Force HTTP/1.1
	fmt.Println("\n[Test 4] Force HTTP/1.1 & TLS 1.2")
	http1Transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
		},
		ForceAttemptHTTP2: false,
	}
	testConnection(endpoint, accessKey, secretKey, bucket, http1Transport)
}

func testConnection(endpoint, accessKey, secretKey, bucket string, transport *http.Transport) {
	ctx := context.TODO()

	var httpClient *http.Client
	if transport != nil {
		httpClient = &http.Client{Transport: transport}
	} else {
		httpClient = http.DefaultClient
	}

	opts := []func(*config.LoadOptions) error{
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion("auto"),
		config.WithHTTPClient(httpClient),
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		return
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	// Try ListObjectsV2 (doesn't upload anything, just checks access)
	fmt.Print("   Listing objects... ")
	_, err = client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int32(1),
	})

	if err != nil {
		fmt.Printf("FAIL: %v\n", err)
	} else {
		fmt.Println("SUCCESS ✅")
	}

	// Try PutObject (upload small file)
	fmt.Print("   Uploading test file... ")
	key := "test-connection.txt"
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader("Hello R2 from Go!"),
	})
	if err != nil {
		fmt.Printf("FAIL: %v\n", err)
	} else {
		fmt.Println("SUCCESS ✅")
		// Clean up
		client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	}
}

func loadEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return // Ignore error if file doesn't exist
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
}
