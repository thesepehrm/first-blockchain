package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// BlockChain is the structure for a blockchain
type BlockChain struct {
	blocks []*Block
}

// Block is the structure of a block in a blockchain
type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
}

// DeriveHash derives the hash for the current block
func (b *Block) DeriveHash() {
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})
	hash := sha256.Sum256(info)
	b.Hash = hash[:]
}

func createBlock(data string, prevHash []byte) *Block {
	block := new(Block)
	block.Data = []byte(data)
	block.PrevHash = prevHash
	block.DeriveHash()
	return block
}

func (chain *BlockChain) addBlock(data string) {
	prevBlock := chain.blocks[len(chain.blocks)-1]
	newBlock := createBlock(data, prevBlock.Hash)
	chain.blocks = append(chain.blocks, newBlock)
}

// Genesis creates the genesis block
func Genesis() *Block {
	return createBlock("Genesis", []byte{})
}

// InitBlockChain makes a new blockchain
func InitBlockChain() *BlockChain {
	return &BlockChain{[]*Block{Genesis()}}
}

func main() {

	chain := InitBlockChain()

	chain.addBlock("#2 Hello!")
	chain.addBlock("#3 World!")

	for _, block := range chain.blocks {
		fmt.Println("-------------------")
		fmt.Printf("Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("PrevHash: %x\n", block.PrevHash)
		fmt.Println("-------------------")
	}
}
