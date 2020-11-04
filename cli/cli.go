package cli

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"gitlab.com/thesepehrm/first-blockchain/blockchain"
	"gitlab.com/thesepehrm/first-blockchain/wallet"
)

// CommandLine is a structure for cli commands
type CommandLine struct{}

func (cli *CommandLine) printHelp() {
	println("Commands:")
	println(" balance -address ADDRESS - Get the balance for the address")
	println(" createchain -address ADDRESS - makes the blockchain and the address mines the genesis")
	println(" send -from ADDRESS -to ADDRESS -amount AMOUNT - Sends some coin from an address to another address")
	println(" print - Prints all of the blocks")
	println("-----Wallets-----")
	println(" createwallet - Creates a new Wallet")
	println(" listaddresses - Lists the addresses of our wallets")
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
		fmt.Println("Transactions:")
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Println("-------------------")
		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) createChain(address string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not valid")
	}

	chain := blockchain.InitBlockChain(address)
	chain.Database.Close()
	fmt.Println("Finished!")
}

func (cli *CommandLine) getBalance(address string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not valid")
	}

	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()

	balance := 0

	decoded := wallet.DecodeBase58([]byte(address))
	pubKeyHash := decoded[1 : len(decoded)-4]
	utxo := chain.FindUTXO(pubKeyHash)
	for _, out := range utxo {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from string, to string, amount int) {
	if !wallet.ValidateAddress(from) {
		log.Panic("Source Address is not valid")
	}
	if !wallet.ValidateAddress(to) {
		log.Panic("Destination Address is not valid")
	}

	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()

	tx := blockchain.NewTransaction(from, to, amount, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})
	fmt.Printf("Transaction #%s was successful!\n", hex.EncodeToString(tx.ID))

}

func (cli *CommandLine) listAddresses() {
	w, err := wallet.CreateWallets()
	wallet.Handle(err)

	addresses := w.GetAllAddresses()
	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CommandLine) createWallet() {
	w, _ := wallet.CreateWallets()
	address := w.AddWallet()
	w.SaveFile()

	fmt.Printf("New wallet address is: %s\n", address)
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

	createWalletCommand := flag.NewFlagSet("createwallet", flag.ExitOnError)

	listAddressesCommand := flag.NewFlagSet("listaddresses", flag.ExitOnError)

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

	case "listaddresses":
		err := listAddressesCommand.Parse(os.Args[2:])
		wallet.Handle(err)

	case "createwallet":
		err := createWalletCommand.Parse(os.Args[2:])
		wallet.Handle(err)

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

	if createWalletCommand.Parsed() {
		cli.createWallet()
	}

	if listAddressesCommand.Parsed() {
		cli.listAddresses()
	}

}
