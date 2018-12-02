package main

import (
	"bytes"
	"context"
	"fmt"

	"google.golang.org/appengine/urlfetch"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type schemaAWSService struct {
	region string
	sess   *session.Session
}

func newSchemaAWSService(ctx context.Context, args struct {
	Region string
}) (*schemaAWSService, error) {
	client := urlfetch.Client(ctx)
	config := &aws.Config{
		Region:      aws.String(args.Region),
		Credentials: credentials.NewEnvCredentials(),
		HTTPClient:  client,
	}

	sess, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}

	service := schemaAWSService{
		region: args.Region,
		sess:   sess,
	}
	return &service, nil
}

// S3 resolved
func (service *schemaAWSService) S3() (*schemaAWSS3Service, error) {
	return newSchemaAWSS3Service(service.sess), nil
}

type schemaAWSS3Service struct {
	svc *s3.S3
}

func newSchemaAWSS3Service(sess *session.Session) *schemaAWSS3Service {
	svc := s3.New(sess)

	service := schemaAWSS3Service{
		svc: svc,
	}
	return &service
}

// Object resolved
func (service *schemaAWSS3Service) Object(args struct {
	Bucket string
	Key    string
}) (*schemaAWSS3Object, error) {
	o := schemaAWSS3Object{
		svc:    service.svc,
		bucket: args.Bucket,
		key:    args.Key,
	}
	return &o, nil
}

type schemaAWSS3Object struct {
	svc    *s3.S3
	bucket string
	key    string
}

func (o *schemaAWSS3Object) Body() (*schemaAWSS3ObjectBody, error) {
	output, err := o.svc.GetObject(&s3.GetObjectInput{
		Bucket: &o.bucket,
		Key:    &o.key,
	})
	if err != nil {
		return nil, fmt.Errorf("Cannot get object '%s'", o.key)
	}

	body := schemaAWSS3ObjectBody{
		output: output,
	}
	return &body, nil
}

type schemaAWSS3ObjectBody struct {
	output *s3.GetObjectOutput
}

// ContentLength resolved
func (o *schemaAWSS3ObjectBody) ContentLength() (*int32, error) {
	if o.output.ContentLength == nil {
		return nil, nil
	}
	l := int32(*o.output.ContentLength)
	return &l, nil
}

// ContentType resolved
func (o *schemaAWSS3ObjectBody) ContentType() (*string, error) {
	return o.output.ContentType, nil
}

// String resolved
func (o *schemaAWSS3ObjectBody) String() (*string, error) {
	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(o.output.Body)
	if err != nil {
		return nil, err
	}
	s := buffer.String()
	return &s, nil
}
