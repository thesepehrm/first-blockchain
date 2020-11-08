package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"runtime"

	"github.com/dgraph-io/badger"
)

const (
	dbPath = "./tmp/blocks_%s"
)

// BlockChain is the structure for a blockchain
type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

// InitBlockChain makes a new blockchain
func InitBlockChain(address, nodeID string) *BlockChain {
	var lastHash []byte

	path := fmt.Sprintf(dbPath, nodeID)

	if DBExists(path) {
		fmt.Println("Blockchain already exists.")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(path)
	opts.Logger = nil

	db, err := openDB(path, opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchains found. Creating a new blockchain...")
			coinbaseTx := CoinbaseTx(address, "Genesis")
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

func ContinueBlockChain(nodeID string) *BlockChain {
	var lastHash []byte

	path := fmt.Sprintf(dbPath, nodeID)

	if !DBExists(path) {
		fmt.Println("No blockchain exists; Make one!")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(path)
	opts.Logger = nil

	db, err := openDB(path, opts)
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

func (chain *BlockChain) getLastBlock() *Block {

	var block *Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		record, err := txn.Get([]byte("lh"))
		Handle(err)

		err = record.Value(func(lastHash []byte) error {
			block, err = chain.GetBlock(lastHash)
			return err
		})
		return err
	})

	Handle(err)
	return block
}

func (chain *BlockChain) GetBlock(blockHash []byte) (*Block, error) {
	var block *Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		blockRecord, err := txn.Get(blockHash)
		if err != nil {
			return err
		}

		err = blockRecord.Value(func(val []byte) error {
			block = Deserialize(val)
			return nil
		})
		return err
	})

	return block, err
}

func (chain *BlockChain) GetBlockHashes() [][]byte {
	var hashes [][]byte

	iter := chain.Iterator()
	for {
		block := iter.Next()

		hashes = append(hashes, block.Hash)

		if len(block.PrevHash) == 0 {
			break
		}

	}
	return hashes
}

func (chain *BlockChain) GetBestHeight() int {
	lastBlock := chain.getLastBlock()
	return lastBlock.Height
}

func (chain *BlockChain) AddBlock(block *Block) {

	err := chain.Database.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(block.Hash); err == nil {
			return nil // block already exists
		}

		err := txn.Set(block.Hash, block.Serialize())
		Handle(err)

		lastBlock := chain.getLastBlock()

		if block.Height > lastBlock.Height {
			err = txn.Set([]byte("lh"), block.Hash)
			chain.LastHash = block.Hash
			Handle(err)

		}
		return nil
	})
	Handle(err)

}

func (chain *BlockChain) MineBlock(transactions []*Transaction) *Block {

	lastBlock := chain.getLastBlock()

	newBlock := CreateBlock(transactions, lastBlock.Hash, lastBlock.Height+1)

	err := chain.Database.Update(func(txn *badger.Txn) error {
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
			if bytes.Equal(tx.ID, ID) {
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
			if !tx.IsCoinbase() {
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
