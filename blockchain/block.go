package blockchain

import (
	"bytes"
	"crypto/sha256"
)

// BlockChain is the structure for a blockchain
type BlockChain struct {
	Blocks []*Block
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

func (chain *BlockChain) AddBlock(data string) {
	prevBlock := chain.Blocks[len(chain.Blocks)-1]
	newBlock := createBlock(data, prevBlock.Hash)
	chain.Blocks = append(chain.Blocks, newBlock)
}

// Genesis creates the genesis block
func Genesis() *Block {
	return createBlock("Genesis", []byte{})
}

// InitBlockChain makes a new blockchain
func InitBlockChain() *BlockChain {
	return &BlockChain{[]*Block{Genesis()}}
}
