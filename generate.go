package main

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"log"
	"os"
)

func uploadTemplateToS3(templatePath string) (string, error) {
	bucket, ok := os.LookupEnv("S3_BUCKET_NAME")
	if !ok {
		return "", errors.New("You must set Environment Variable 'S3_BUCKET_NAME'")
	}
	key, ok := os.LookupEnv("AWS_ACCESS_KEY")
	if !ok {
		return "", errors.New("You must set Environment Variable 'AWS_ACCESS_KEY'")
	}

	template, err := os.Open(templatePath)
	if err != nil {
		return "", err
	}
	defer template.Close()

	uploader := s3manager.NewUploader(
		session.New(aws.NewConfig().WithRegion("ap-northeast-1")), nil,
	)
	result, err := uploader.Upload(
		&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   template,
		},
	)
	if err != nil {
		return "", err
	}

	return result.Location, nil
}

func createAnsibleTargets(stackName, templateURL *string) error {
	client := cloudformation.New(
		session.New(aws.NewConfig().WithRegion("ap-northeast-1")), nil,
	)

	_, err := client.CreateStack(
		&cloudformation.CreateStackInput{
			StackName:   stackName,
			TemplateURL: templateURL,
		},
	)
	if err != nil {
		return err
	}

	err = client.WaitUntilStackCreateComplete(
		&cloudformation.DescribeStacksInput{StackName: stackName},
	)
	if err != nil {
		return err
	}

	return nil
}

func printInventory(stackName string) error {
	return nil
}

func main() {
	const (
		stackName    = "AnsibleTargets"
		templatePath = "./ec2.yml"
	)

	templateURL, err := uploadTemplateToS3(templatePath)
	if err != nil {
		log.Fatal(err)
	}

	err = createAnsibleTargets(aws.String(stackName), aws.String(templateURL))
	if err != nil {
		log.Fatal(err)
	}

	if err := printInventory(stackName); err != nil {
		log.Fatal(err)
	}
	return
}
