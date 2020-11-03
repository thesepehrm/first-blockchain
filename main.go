package main

import (
	"os"

	"gitlab.com/thesepehrm/first-blockchain/cli"
)

func main() {
	defer os.Exit(0)

	cli := cli.CommandLine{}
	cli.Run()

}
