package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
)

// Block is the structure of a block in a blockchain
type Block struct {
	Nonce        int
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash := sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

func CreateBlock(txns []*Transaction, prevHash []byte) *Block {
	block := new(Block)
	block.Transactions = txns
	block.PrevHash = prevHash
	pow := NewProof(block)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash

	return block
}

// Genesis creates the genesis block
func Genesis(coinBase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinBase}, []byte{})
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
