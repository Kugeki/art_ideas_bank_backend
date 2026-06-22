package s3client

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"io"
)

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Provider interface {
	ProvideS3Client(ctx context.Context) (*s3.Client, error)
	GetBucket() string
}

type S3Client struct {
	Client *s3.Client
	Bucket string
}

func New(ctx context.Context, provider Provider) (*S3Client, error) {
	client, err := provider.ProvideS3Client(ctx)
	if err != nil {
		return nil, err
	}

	return &S3Client{
		Client: client,
		Bucket: provider.GetBucket(),
	}, nil
}

func (c *S3Client) Test(ctx context.Context) (*s3.ListBucketsOutput, error) {
	result, err := c.Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *S3Client) Upload(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, err := c.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.Bucket),
		Key:         aws.String(fmt.Sprintf("%v/%v", c.Bucket, key)),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *S3Client) Download(ctx context.Context, key string) (io.ReadCloser, string, error) {
	result, err := c.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(fmt.Sprintf("%v/%v", c.Bucket, key)),
	})
	if err != nil {
		return nil, "", err
	}
	return result.Body, *result.ContentType, nil
}

func (c *S3Client) Delete(ctx context.Context, key string) error {
	_, err := c.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(fmt.Sprintf("%v/%v", c.Bucket, key)),
	})
	if err != nil {
		return err
	}
	return nil
}
