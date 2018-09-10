package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
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
		session.New(aws.NewConfig().WithRegion("ap-northeast-1"), nil),
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
		session.New(aws.NewConfig().WithRegion("ap-northeast-1"), nil),
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
	client := ec2.New(
		session.New(aws.NewConfig().WithRegion("ap-northeast-1"), nil),
	)
	description, err := client.DescribeInstances(
		&ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("tag:aws:cloudformation:stack-name"),
					Values: []*string{aws.String(stackName)},
				},
			},
		},
	)
	if err != nil {
		return err
	}

	inventoryGroupMembers := make(map[string][]string)
	var keyname string

	for _, reservation := range description.Reservations {
		instance := reservation.Instances[0]

		// Always use same private key among instances defined in same template
		keyname = *instance.KeyName
		var groupName string
		for _, tag := range instance.Tags {
			if *tag.Key == "AnsibleInventoryGroup" {
				groupName = *tag.Value
			}
		}
		publicIp := *instance.PublicIpAddress
		if _, exists := inventoryGroupMembers[groupName]; exists {
			inventoryGroupMembers[groupName] = append(inventoryGroupMembers[groupName], publicIp)
		} else {
			inventoryGroupMembers[groupName] = []string{publicIp}
		}
	}

	for group, members := range inventoryGroupMembers {
		fmt.Printf("[%s]\n", group)
		for _, ip := range members {
			fmt.Printf("%s\n", ip)
		}
	}
	fmt.Printf("[all:vars]\n")
	fmt.Printf("ansible_ssh_user=ec2-user\n")
	fmt.Printf("ansible_ssh_private_key_file=~/.ssh/%s.pem\n", keyname)
	return nil
}

func main() {
	const (
		stackName    = "AnsibleTargets03"
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