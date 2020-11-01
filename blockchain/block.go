package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

// Block is the structure of a block in a blockchain
type Block struct {
	Nonce    int
	Hash     []byte
	Data     []byte
	PrevHash []byte
}

func createBlock(data string, prevHash []byte) *Block {
	block := new(Block)
	block.Data = []byte(data)
	block.PrevHash = prevHash
	pow := NewProof(block)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash

	return block
}

// Genesis creates the genesis block
func Genesis() *Block {
	return createBlock("Genesis", []byte{})
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(b)

	Handle(err)

	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)

	Handle(err)

	return &block

}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
