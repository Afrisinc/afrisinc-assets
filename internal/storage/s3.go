package storage

// S3Store implements Store for any S3-compatible object store
// (AWS S3, MinIO, Backblaze B2, Cloudflare R2, DigitalOcean Spaces).
//
// Wire it up by importing the AWS SDK v2:
//
//   go get github.com/aws/aws-sdk-go-v2/...
//
// and filling in the struct fields from your StorageConfig.
// The interface contract is identical to LocalStore — swap the backend
// in server.New() without touching any handler or service code.
//
// Minimal implementation sketch (not compiled, for reference):
//
//   import (
//       "github.com/aws/aws-sdk-go-v2/aws"
//       "github.com/aws/aws-sdk-go-v2/config"
//       "github.com/aws/aws-sdk-go-v2/credentials"
//       "github.com/aws/aws-sdk-go-v2/service/s3"
//   )
//
//   type S3Store struct {
//       client *s3.Client
//       bucket string
//   }
//
//   func NewS3(cfg *appconfig.StorageConfig) (*S3Store, error) {
//       resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
//           return aws.Endpoint{URL: cfg.S3Endpoint}, nil
//       })
//       awsCfg, _ := config.LoadDefaultConfig(context.Background(),
//           config.WithEndpointResolverWithOptions(resolver),
//           config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
//               cfg.S3AccessKey, cfg.S3SecretKey, "",
//           )),
//           config.WithRegion(cfg.S3Region),
//       )
//       return &S3Store{client: s3.NewFromConfig(awsCfg), bucket: cfg.S3Bucket}, nil
//   }
//
// All five interface methods (Put/Get/Delete/Exists/URL) map directly to
// PutObject, GetObject, DeleteObject, HeadObject, and PresignGetObject.

// Placeholder type so the package compiles before the full implementation is added.
type S3Store struct{}
