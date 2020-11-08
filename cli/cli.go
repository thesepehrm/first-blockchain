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
	"gitlab.com/thesepehrm/first-blockchain/network"
	"gitlab.com/thesepehrm/first-blockchain/wallet"
)

// CommandLine is a structure for cli commands
type CommandLine struct{}

func (cli *CommandLine) printHelp() {
	println("Commands:")
	println(" startnode [-miner] ADDRESS - Starts a node, -miner flag sets the node to be a miner")
	println(" balance -address ADDRESS - Get the balance for the address")
	println(" createchain -address ADDRESS - makes the blockchain and the address mines the genesis")
	println(" send -from ADDRESS -to ADDRESS -amount AMOUNT - Sends some coin from an address to another address")
	println(" print - Prints all of the blocks")
	println("reindexutxo nodeID- Rebuilds the utxo database")
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

func (cli *CommandLine) startNode(nodeID, minerAddress string) {
	fmt.Printf("Starting Node %s\n", nodeID)
	if len(minerAddress) > 0 {
		if wallet.ValidateAddress(minerAddress) {
			fmt.Printf("Node is a miner, wallet address for rewards: %s\n", minerAddress)
		} else {
			log.Panic("Miner address is not valid!")
		}
	}

	network.Start(nodeID, minerAddress)

}

func (cli *CommandLine) printChain(nodeID string) {
	chain := blockchain.ContinueBlockChain(nodeID)
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

func (cli *CommandLine) createChain(address, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not valid")
	}

	chain := blockchain.InitBlockChain(address, nodeID)
	chain.Database.Close()
	fmt.Println("Finished!")
}

func (cli *CommandLine) getBalance(address string, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not valid")
	}

	chain := blockchain.ContinueBlockChain(nodeID)
	UTXO := blockchain.UTXOSet{Blockchain: chain}
	defer chain.Database.Close()

	balance := 0

	decoded := wallet.DecodeBase58([]byte(address))
	pubKeyHash := decoded[1 : len(decoded)-4]
	unspentTxs := UTXO.FindUnspentTransactions(pubKeyHash)

	for _, out := range unspentTxs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from string, to string, amount int, nodeID string) {
	if !wallet.ValidateAddress(from) {
		log.Panic("Source Address is not valid")
	}
	if !wallet.ValidateAddress(to) {
		log.Panic("Destination Address is not valid")
	}

	chain := blockchain.ContinueBlockChain(nodeID)
	UTXO := blockchain.UTXOSet{Blockchain: chain}
	defer chain.Database.Close()
	tx := blockchain.NewTransaction(from, to, amount, &UTXO, nodeID)
	//chain.MineBlock([]*blockchain.Transaction{tx})
	network.SendTX(network.KnownNodes[0], tx)
	fmt.Printf("Sent Transaction #%s\n", hex.EncodeToString(tx.ID))

}

func (cli *CommandLine) reIndexUTXO(nodeID string) {
	chain := blockchain.ContinueBlockChain(nodeID)
	UTXO := blockchain.UTXOSet{Blockchain: chain}
	defer chain.Database.Close()

	UTXO.ReIndex()
	count := UTXO.CountTransactions()
	fmt.Printf("Done! There are %d UTXOs in the database\n", count)
}

func (cli *CommandLine) listAddresses(nodeID string) {
	w, err := wallet.CreateWallets(nodeID)
	wallet.Handle(err)

	addresses := w.GetAllAddresses()
	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CommandLine) createWallet(nodeID string) {
	w, _ := wallet.CreateWallets(nodeID)
	address := w.AddWallet()
	w.SaveFile(nodeID)

	fmt.Printf("New wallet address is: %s\n", address)
}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		log.Panic("NODE_ID env is not set!")
		runtime.Goexit()
	}

	createChainCommand := flag.NewFlagSet("createchain", flag.ExitOnError)
	createChainData := createChainCommand.String("address", "", "Address of the miner of the genesis")

	balanceCommand := flag.NewFlagSet("balance", flag.ExitOnError)
	balanceData := balanceCommand.String("address", "", "Address of the wallet")

	sendCommand := flag.NewFlagSet("send", flag.ExitOnError)
	sendFrom := sendCommand.String("from", "", "Source wallet address")
	sendTo := sendCommand.String("to", "", "Destination wallet address")
	sendAmount := sendCommand.Int("amount", 0, "Transfer amount")

	printChainCommand := flag.NewFlagSet("print", flag.ExitOnError)

	reIndexUTXOCommand := flag.NewFlagSet("reindexutxo", flag.ExitOnError)

	createWalletCommand := flag.NewFlagSet("createwallet", flag.ExitOnError)

	listAddressesCommand := flag.NewFlagSet("listaddresses", flag.ExitOnError)

	startNodeCommand := flag.NewFlagSet("startnode", flag.ExitOnError)
	startNodeData := startNodeCommand.String("miner", "", "Enables mining and requires an address for the rewards")

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
	case "reindexutxo":
		err := reIndexUTXOCommand.Parse(os.Args[2:])
		blockchain.Handle(err)
	case "startnode":
		err := startNodeCommand.Parse(os.Args[2:])
		network.Handle(err)

	default:
		cli.printHelp()
		runtime.Goexit()
	}

	if createChainCommand.Parsed() {
		if *createChainData == "" {
			createChainCommand.Usage()
			runtime.Goexit()
		}
		cli.createChain(*createChainData, nodeID)
	}

	if balanceCommand.Parsed() {
		if *balanceData == "" {
			balanceCommand.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*balanceData, nodeID)
	}

	if sendCommand.Parsed() {
		if *sendAmount == 0 && *sendFrom == "" && *sendTo == "" {
			sendCommand.Usage()
			runtime.Goexit()
		}
		cli.send(*sendFrom, *sendTo, *sendAmount, nodeID)
	}

	if printChainCommand.Parsed() {
		cli.printChain(nodeID)
	}

	if createWalletCommand.Parsed() {
		cli.createWallet(nodeID)
	}

	if listAddressesCommand.Parsed() {
		cli.listAddresses(nodeID)
	}

	if reIndexUTXOCommand.Parsed() {
		cli.reIndexUTXO(nodeID)
	}

	if startNodeCommand.Parsed() {
		cli.startNode(nodeID, *startNodeData)
	}

}
