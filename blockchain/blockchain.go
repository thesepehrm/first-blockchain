package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "Coinbase transaction"
)

// BlockChain is the structure for a blockchain
type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

func DBExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// InitBlockChain makes a new blockchain
func InitBlockChain(address string) *BlockChain {
	var lastHash []byte

	if DBExists() {
		fmt.Println("Blockchain already exists.")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchains found. Creating a new blockchain...")
			coinbaseTx := CoinbaseTx(address, genesisData)
			genesis := Genesis(coinbaseTx)
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

func ContinueBlockChain(address string) *BlockChain {
	var lastHash []byte

	if DBExists() == false {
		fmt.Println("No blockchain exists; Make one!")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		record, err := txn.Get([]byte("lh"))
		Handle(err)
		err = record.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})
		return err
	})

	Handle(err)

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

func (chain *BlockChain) AddBlock(transactions []*Transaction) *Block {
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

	newBlock := CreateBlock(transactions, lastHash)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)
		chain.LastHash = newBlock.Hash
		return err
	})

	Handle(err)
	return newBlock
}

func (chain *BlockChain) FindTransactions(ID []byte) (Transaction, error) {
	iterator := chain.Iterator()
	for {
		block := iterator.Next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("Transaction not found")
}

func (chain *BlockChain) FindUTXO() map[string]TxOutputs {
	UTXO := make(map[string]TxOutputs)
	spentTxs := make(map[string][]int)

	iter := chain.Iterator()
	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTxs[txID] != nil {

					for _, spentOut := range spentTxs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTxs[inTxID] = append(spentTxs[inTxID], in.Out)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return UTXO
}

func (chain *BlockChain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey) {
	prevTxs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTx, err := chain.FindTransactions(in.ID)
		Handle(err)
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}
	tx.Sign(privateKey, prevTxs)

}

func (chain *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTxs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTx, err := chain.FindTransactions(in.ID)
		Handle(err)
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}

	return tx.Verify(prevTxs)
}
