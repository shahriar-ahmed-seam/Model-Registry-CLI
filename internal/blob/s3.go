// Package blob provides the BlobStore interface for storing and retrieving
// model blobs, with an S3-compatible implementation.
package blob

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"

	"ml-reg/internal/config"
)

// s3Store implements the BlobStore interface using S3-compatible object storage.
type s3Store struct {
	client *s3.Client
	bucket string
}

// NewS3Store creates a new S3-compatible blob store based on the provided config.
// The config must contain an endpoint, bucket name, and optional region.
// It uses path-style addressing for MinIO compatibility.
func NewS3Store(cfg *config.Config) (BlobStore, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is required for S3 store")
	}
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("bucket is required for S3 store")
	}

	// Create custom endpoint resolver
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID {
			return aws.Endpoint{
				URL:               cfg.Endpoint,
				SigningRegion:     cfg.Region,
				HostnameImmutable: true,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	// Resolve credentials via the AWS default credential chain. Secrets are
	// never read from the config file. cfg.CredsRef, when set, is treated as a
	// non-secret reference to an AWS shared-config profile name; otherwise the
	// SDK resolves credentials from the environment (AWS_ACCESS_KEY_ID /
	// AWS_SECRET_ACCESS_KEY), the shared credentials file, or other providers in
	// the default chain.
	loadOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithEndpointResolverWithOptions(customResolver),
	}
	if cfg.CredsRef != "" {
		loadOpts = append(loadOpts, awsconfig.WithSharedConfigProfile(cfg.CredsRef))
	}

	// Build AWS config with customizations
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	// Create S3 client with custom options. Credentials come from the resolved
	// awsCfg (default chain / named profile), not from any config secret.
	clientOptions := func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true // MinIO compatibility
	}

	client := s3.NewFromConfig(awsCfg, clientOptions)

	return &s3Store{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// Exists reports whether an object with the given content-hash key is present in S3.
func (s *s3Store) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			// Check if it's a NotFound error
			if strings.Contains(apiErr.ErrorCode(), "NotFound") {
				return false, nil
			}
		}
		// For other errors (network, permissions, etc.), return the error
		return false, fmt.Errorf("failed to check existence of key %q: %w", key, err)
	}

	return true, nil
}

// Put streams content to S3 under the given key using multipart upload for large files.
func (s *s3Store) Put(ctx context.Context, key string, r io.Reader, size int64) error {
	uploader := manager.NewUploader(s.client)

	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   r,
	})

	if err != nil {
		return fmt.Errorf("failed to upload key %q to S3: %w", key, err)
	}

	return nil
}

// Get streams the object content from S3 to the writer.
func (s *s3Store) Get(ctx context.Context, key string, w io.Writer) error {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if strings.Contains(apiErr.ErrorCode(), "NotFound") {
				return fmt.Errorf("%w: %s", ErrNotFound, key)
			}
		}
		return fmt.Errorf("failed to get key %q from S3: %w", key, err)
	}
	defer result.Body.Close()

	// Stream the object body to the writer
	_, err = io.Copy(w, result.Body)
	if err != nil {
		return fmt.Errorf("failed to stream S3 object to writer for key %q: %w", key, err)
	}

	return nil
}