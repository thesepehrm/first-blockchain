package cli

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"gitlab.com/thesepehrm/first-blockchain/blockchain"
)

// CommandLine is a structure for cli commands
type CommandLine struct{}

func (cli *CommandLine) printHelp() {
	println("Commands:")
	println(" balance -address ADDRESS - Get the balance for the address")
	println(" createchain -address ADDRESS - makes the blockchain and the address mines the genesis")
	println(" send -from ADDRESS -to ADDRESS -amount AMOUNT - Sends some coin from an address to another address")
	println(" print - Prints all of the blocks")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printHelp()
		runtime.Goexit()
	}
}

func (cli *CommandLine) printChain() {
	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()

	iter := chain.Iterator()
	for {
		block := iter.Next()
		fmt.Println("-------------------")
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println("-------------------")
		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) createChain(address string) {
	chain := blockchain.InitBlockChain(address)
	chain.Database.Close()
	fmt.Println("Finished!\n")
}

func (cli *CommandLine) getBalance(address string) {
	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()

	balance := 0
	utxo := chain.FindUTXO(address)
	for _, out := range utxo {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from string, to string, amount int) {
	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()

	tx := blockchain.NewTransaction(from, to, amount, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})
	fmt.Printf("Transaction #%s was successful!\n", hex.EncodeToString(tx.ID))

}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	createChainCommand := flag.NewFlagSet("createchain", flag.ExitOnError)
	createChainData := createChainCommand.String("address", "", "Address of the miner of the genesis")

	balanceCommand := flag.NewFlagSet("balance", flag.ExitOnError)
	balanceData := balanceCommand.String("address", "", "Address of the wallet")

	sendCommand := flag.NewFlagSet("send", flag.ExitOnError)
	sendFrom := sendCommand.String("from", "", "Source wallet address")
	sendTo := sendCommand.String("to", "", "Destination wallet address")
	sendAmount := sendCommand.Int("amount", 0, "Transfer amount")

	printChainCommand := flag.NewFlagSet("print", flag.ExitOnError)

	switch os.Args[1] {
	case "createchain":
		err := createChainCommand.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "balance":
		err := balanceCommand.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "send":
		err := sendCommand.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "print":
		err := printChainCommand.Parse(os.Args[2:])
		blockchain.Handle(err)
	default:
		cli.printHelp()
		runtime.Goexit()
	}

	if createChainCommand.Parsed() {
		if *createChainData == "" {
			createChainCommand.Usage()
			runtime.Goexit()
		}
		cli.createChain(*createChainData)
	}

	if balanceCommand.Parsed() {
		if *balanceData == "" {
			balanceCommand.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*balanceData)
	}

	if sendCommand.Parsed() {
		if *sendAmount == 0 && *sendFrom == "" && *sendTo == "" {
			sendCommand.Usage()
			runtime.Goexit()
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if printChainCommand.Parsed() {
		cli.printChain()
	}

}
