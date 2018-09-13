ec2inv
====

Generator of Ansible's inventory for EC2 instances

## Description

When you need Amazon EC2 instances provisioned by Ansible, this tool may help you.

It aggregates descriptions of EC2 instances which belong to specified stack and
shows what you should write to your inventory file.

## Requirements
* Go

## Usage
```
$ ec2inv --help
ec2inv shows Ansible's inventory for EC2 instances

Usage:
  ec2inv [flags]

Flags:
  -h, --help                             help for ec2inv
  -i, --inventory-group-tag-key string   Tag key attached to EC2 instances to specify inventory group
  -s, --stack-name string                Name of a stack which EC2 instances belong to
      --version                          version for ec2inv
```

As you can see, ec2inv requires stack name and tag key.

This means that EC2 instances to be provisioned must be created by AWS CloudFormation
and each instance must have additional tag to determine inventory group.

See [sample.yml](https://github.com/it-akumi/ec2inv/blob/master/sample.yml).

## Install
```
$ go get -u github.com/it-akumi/ec2inv
```

## Author
[Takumi Ishii](https://github.com/it-akumi)

## License
[MIT](https://github.com/it-akumi/ec2inv/blob/master/LICENSE)
