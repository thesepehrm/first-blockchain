package main

import (
	"fmt"

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
		fmt.Printf("PrevHash: %x\n", block.PrevHash)
		fmt.Println("-------------------")
	}
}
