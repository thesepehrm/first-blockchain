package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"gitlab.com/thesepehrm/first-blockchain/blockchain"
)

// CommandLine is a structure for cli commands
type CommandLine struct {
	blockchain *blockchain.BlockChain
}

func (cli *CommandLine) printHelp() {
	println("Commands:")
	println(" add -block BLOCK_DATA - Add a block to the chain")
	println(" print - Prints all of the blocks")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printHelp()
		runtime.Goexit()
	}
}

func (cli *CommandLine) addBlock(data string) {
	cli.blockchain.AddBlock(data)
	fmt.Println("Created a new block!")
}

func (cli *CommandLine) printChain() {
	iter := cli.blockchain.Iterator()

	for {
		block := iter.Next()
		fmt.Println("-------------------")
		fmt.Printf("Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println("-------------------")
		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) run() {
	cli.validateArgs()

	addBlockCommand := flag.NewFlagSet("add", flag.ExitOnError)
	addBlockData := addBlockCommand.String("block", "", "Block data")

	printChainCommand := flag.NewFlagSet("print", flag.ExitOnError)

	switch os.Args[1] {
	case "add":
		err := addBlockCommand.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "print":
		err := printChainCommand.Parse(os.Args[2:])
		blockchain.Handle(err)
	default:
		cli.printHelp()
		runtime.Goexit()
	}

	if addBlockCommand.Parsed() {
		if *addBlockData == "" {
			addBlockCommand.Usage()
			runtime.Goexit()
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCommand.Parsed() {
		cli.printChain()
	}

}

func main() {
	defer os.Exit(0)
	chain := blockchain.InitBlockChain()
	defer chain.Database.Close()

	cli := CommandLine{chain}
	cli.run()

}
