package main

import (
	"bytes"
	"context"
	"fmt"

	"google.golang.org/appengine/urlfetch"

	"github.com/BurntSushi/toml"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// ParseAWSCommand parses a /aws … command
func ParseAWSCommand(subcommands []string, params string) (Command, error) {
	if len(subcommands) == 1 {
		if subcommands[0] == "s3" {
			return ParseAWSS3Command(params)
		}
	} else if len(subcommands) == 2 {
		if subcommands[0] == "s3" && subcommands[1] == "object" {
			return ParseAWSS3ObjectCommand(params)
		}
	}

	return nil, fmt.Errorf("Unknown subcommand(s) %v", subcommands)
}

// A AWSS3Command represents the `/aws s3` command
type AWSS3Command struct {
	Bucket string `toml:"bucket"`
	Region string `toml:"region"`
}

// Subcommands resolved
func (cmd *AWSS3Command) Subcommands() *[]string {
	return nil
}

// Params resolved
func (cmd *AWSS3Command) Params() *CommandParams {
	return nil
}

// ParseAWSS3Command creates a new `/aws {input}` command
func ParseAWSS3Command(params string) (*AWSS3Command, error) {
	var cmd AWSS3Command

	_, err := toml.Decode(params, &cmd)
	if err != nil {
		return nil, err
	}

	return &cmd, nil
}

// AWSCommandResult is named the same in GraphQL
type AWSCommandResult struct {
}

// Run converts the S3 to a preview
func (cmd *AWSS3Command) Run(ctx context.Context) (CommandResult, error) {
	client := urlfetch.Client(ctx)
	config := &aws.Config{
		Region:      aws.String(cmd.Region),
		Credentials: credentials.NewEnvCredentials(),
		HTTPClient:  client,
	}

	sess, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)

	output, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: &cmd.Bucket,
	})
	if err != nil {
		return nil, fmt.Errorf("Cannot list objects %v", err)
	}

	var htmlBuffer bytes.Buffer
	htmlBuffer.WriteString(`<pre>`)

	for _, o := range output.Contents {
		htmlBuffer.WriteString(*o.Key)
		htmlBuffer.WriteString("<br>")
		// fmt.Printf("* %s created on %s\n", aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}
	htmlBuffer.WriteString(`</pre>`)

	result := DangerousHTMLCommandResultFromSafe(htmlBuffer.String())

	return result, nil
}

// A AWSS3ObjectCommand represents the `/aws s3` command
type AWSS3ObjectCommand struct {
	Bucket string `toml:"bucket"`
	Region string `toml:"region"`
	Key    string `toml:"key"`
}

// Subcommands resolved
func (cmd *AWSS3ObjectCommand) Subcommands() *[]string {
	return nil
}

// Params resolved
func (cmd *AWSS3ObjectCommand) Params() *CommandParams {
	return nil
}

// ParseAWSS3ObjectCommand creates a new `/aws {input}` command
func ParseAWSS3ObjectCommand(params string) (*AWSS3ObjectCommand, error) {
	var cmd AWSS3ObjectCommand

	_, err := toml.Decode(params, &cmd)
	if err != nil {
		return nil, err
	}

	return &cmd, nil
}

// Run converts the S3 to a preview
func (cmd *AWSS3ObjectCommand) Run(ctx context.Context) (CommandResult, error) {
	client := urlfetch.Client(ctx)
	config := &aws.Config{
		Region:      aws.String(cmd.Region),
		Credentials: credentials.NewEnvCredentials(),
		HTTPClient:  client,
	}

	sess, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)

	output, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: &cmd.Bucket,
		Key:    &cmd.Key,
	})
	if err != nil {
		return nil, fmt.Errorf("Cannot get object '%s'", cmd.Key)
	}

	var htmlBuffer bytes.Buffer
	htmlBuffer.WriteString(`<pre>`)
	htmlBuffer.ReadFrom(output.Body)
	htmlBuffer.WriteString(`</pre>`)

	result := DangerousHTMLCommandResultFromSafe(htmlBuffer.String())

	return result, nil
}
