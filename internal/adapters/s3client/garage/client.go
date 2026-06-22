package garage

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	transport "github.com/aws/smithy-go/endpoints"
	"net/url"
)

type S3Provider struct {
	Endpoint  string
	Region    string
	AccessKey string
	SecretKey string
	Bucket    string
}

type BucketError struct {
	Bucket string
	Err    error
}

func (b *BucketError) Error() string {
	return fmt.Sprintf("bucket \"%v\": %s", b.Bucket, b.Err.Error())
}

func (b *BucketError) Unwrap() error {
	return b.Err
}

var (
	ErrNoBucket = errors.New("s3 don't have provided bucket")
)

func (p *S3Provider) ProvideS3Client(ctx context.Context) (*s3.Client, error) {
	sdkConfig, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(p.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(p.AccessKey, p.SecretKey, "")),
	)

	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(sdkConfig, func(o *s3.Options) {
		o.UsePathStyle = true
		o.EndpointResolverV2 = &staticEndpointResolver{
			endpoint: p.Endpoint,
			region:   p.Region,
		}
	})

	result, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	found := false
	if result != nil {
		for _, bucket := range result.Buckets {
			if *bucket.Name == p.Bucket {
				found = true
				break
			}
		}
	}

	if !found {
		return nil, &BucketError{
			Bucket: p.Bucket,
			Err:    ErrNoBucket,
		}
	}

	return client, nil
}

func (p *S3Provider) GetBucket() string {
	return p.Bucket
}

type staticEndpointResolver struct {
	endpoint string
	region   string
}

func (r *staticEndpointResolver) ResolveEndpoint(
	ctx context.Context, params s3.EndpointParameters,
) (ep transport.Endpoint, err error) {
	u, err := url.Parse(r.endpoint)
	if err != nil {
		return ep, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	return transport.Endpoint{
		URI: *u,
		Properties: func() smithy.Properties {
			var props smithy.Properties
			props.Set("HostnameImmutable", true)
			return props
		}(),
	}, nil
}
