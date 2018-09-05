package main

import (
	"log"
)

func createAnsibleTargets(stackName string) error {
	return nil
}

func printInventory(stackName string) error {
	return nil
}

func main() {
	const (
		stackName = "AnsibleTargets"
	)

	if err := createAnsibleTargets(stackName); err != nil {
		log.Fatal(err)
	}

	if err := printInventory(stackName); err != nil {
		log.Fatal(err)
	}
	return
}
