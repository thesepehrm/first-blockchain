package blockchain

import (
	"fmt"

	"github.com/dgraph-io/badger"
)

const (
	dbPath = "./tmp/blocks"
)

// BlockChain is the structure for a blockchain
type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

// InitBlockChain makes a new blockchain
func InitBlockChain() *BlockChain {
	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchains found. Creating a new blockchain...")
			genesis := Genesis()
			fmt.Println("Genesis Created!")
			err = txn.Set(genesis.Hash, genesis.Serialize())
			Handle(err)
			err = txn.Set([]byte("lh"), genesis.Hash)

			lastHash = genesis.Hash
			return err
		} else {
			record, err := txn.Get([]byte("lh"))
			Handle(err)
			err = record.Value(func(val []byte) error {
				lastHash = append([]byte{}, val...)
				return nil
			})
			return err
		}
	})

	Handle(err)

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

func (chain *BlockChain) AddBlock(data string) {
	var lastHash []byte
	err := chain.Database.View(func(txn *badger.Txn) error {
		record, err := txn.Get([]byte("lh"))
		Handle(err)
		err = record.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})
		return err
	})
	Handle(err)

	newBlock := createBlock(data, lastHash)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)
		chain.LastHash = newBlock.Hash
		return err
	})

	Handle(err)
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := BlockChainIterator{chain.LastHash, chain.Database}

	return &iter
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		record, err := txn.Get(iter.CurrentHash)
		Handle(err)

		var serializedBlock []byte
		err = record.Value(func(val []byte) error {
			serializedBlock = append([]byte{}, val...)
			return nil
		})
		Handle(err)
		block = Deserialize(serializedBlock)
		return err
	})
	Handle(err)

	iter.CurrentHash = block.PrevHash

	return block
}
