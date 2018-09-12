package command

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
	"os"
)

// flags
var (
	stackName            string
	inventoryGroupTagKey string
)

func init() {
	rootCmd.Flags().StringVarP(
		&stackName,
		"stack-name",
		"s",
		"",
		"Name of a stack which EC2 instances belong to",
	)
	rootCmd.Flags().StringVarP(
		&inventoryGroupTagKey,
		"inventory-group-tag-key",
		"i",
		"",
		"Tag key attached to EC2 instances to specify inventory group",
	)

	rootCmd.MarkFlagRequired("stack-name")
	rootCmd.MarkFlagRequired("inventory-group-tag-key")
}

var rootCmd = &cobra.Command{
	Use:     "ec2inv",
	Short:   "ec2inv shows Ansible's inventory for EC2 instances",
	Version: "0.0",
	RunE: func(cmd *cobra.Command, args []string) error {
		classifiedInstances, keyname, err := classifyEC2instances(stackName)
		if err != nil {
			return err
		}
		printInventory(classifiedInstances, keyname)
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type inventoryGroupMembers map[string][]string

func classifyEC2instances(stackName string) (inventoryGroupMembers, string, error) {
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
		return nil, "", err
	}

	classifiedInstances := make(inventoryGroupMembers)
	var keyname string

	for _, reservation := range description.Reservations {
		instance := reservation.Instances[0]

		// Always use same private key among instances defined in same template
		keyname = *instance.KeyName
		var groupName string
		for _, tag := range instance.Tags {
			if *tag.Key == inventoryGroupTagKey {
				groupName = *tag.Value
			}
		}
		publicIp := *instance.PublicIpAddress
		if _, exists := classifiedInstances[groupName]; exists {
			classifiedInstances[groupName] = append(classifiedInstances[groupName], publicIp)
		} else {
			classifiedInstances[groupName] = []string{publicIp}
		}
	}

	return classifiedInstances, keyname, nil
}

func printInventory(classifiedInstances inventoryGroupMembers, keyname string) {
	for group, members := range classifiedInstances {
		fmt.Printf("[%s]\n", group)
		for _, ip := range members {
			fmt.Printf("%s\n", ip)
		}
	}
	fmt.Printf("[all:vars]\n")
	fmt.Printf("ansible_ssh_user=ec2-user\n")
	fmt.Printf("ansible_ssh_private_key_file=~/.ssh/%s.pem\n", keyname)
	return
}
