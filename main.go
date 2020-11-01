package main

import (
	"fmt"
	"strconv"

	"gitlab.com/thesepehrm/first-blockchain/blockchain"
)

func main() {

	chain := blockchain.InitBlockChain()

	chain.AddBlock("#2 Hello!")
	chain.AddBlock("#3 World!")

	for _, block := range chain.Blocks {
		fmt.Println("-------------------")
		fmt.Printf("Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := blockchain.NewProof(block)
		fmt.Println("Proof of work:")
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println("-------------------")
	}
}
