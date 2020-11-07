package blockchain

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/dgraph-io/badger"
)

var (
	utxoPrefix = []byte("utxo-")
	//utxoPrefixLength = len(utxoPrefix)
)

type UTXOSet struct {
	Blockchain *BlockChain
}

func (u *UTXOSet) DeleteKeysByPrefix(prefix []byte) error {
	err := u.Blockchain.Database.DropPrefix(prefix)
	return err
}

func (u *UTXOSet) CountTransactions() int {
	db := u.Blockchain.Database

	count := 0
	err := db.View(func(txn *badger.Txn) error {
		iterator := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iterator.Close()

		for iterator.Seek(utxoPrefix); iterator.ValidForPrefix(utxoPrefix); iterator.Next() {
			count++
		}
		return nil
	})

	Handle(err)
	return count
}

func (u *UTXOSet) FindUnspentTransactions(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput

	db := u.Blockchain.Database

	err := db.View(func(txn *badger.Txn) error {
		iterator := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iterator.Close()
		for iterator.Seek(utxoPrefix); iterator.ValidForPrefix(utxoPrefix); iterator.Next() {
			var outputs TxOutputs
			err := iterator.Item().Value(func(val []byte) error {
				outputs = DeserializeOutputs(val)
				return nil
			})
			Handle(err)

			for _, out := range outputs.Outputs {
				if out.isLockedWith(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}

		}
		return nil
	})
	Handle(err)

	return UTXOs
}

func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	accumulated := 0

	db := u.Blockchain.Database

	err := db.View(func(txn *badger.Txn) error {
		iterator := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iterator.Close()
		for iterator.Seek(utxoPrefix); iterator.ValidForPrefix(utxoPrefix); iterator.Next() {
			var outputs TxOutputs

			txID := hex.EncodeToString(bytes.TrimPrefix(iterator.Item().Key(), utxoPrefix))

			err := iterator.Item().Value(func(val []byte) error {
				outputs = DeserializeOutputs(val)
				return nil
			})

			Handle(err)

			for outID, out := range outputs.Outputs {
				if out.isLockedWith(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOuts[txID] = append(unspentOuts[txID], outID)

					if accumulated >= out.Value {
						break
					}
				}

			}

		}

		return nil
	})
	Handle(err)

	return accumulated, unspentOuts

}

func (u *UTXOSet) Update(block *Block) {
	db := u.Blockchain.Database

	err := db.Update(func(txn *badger.Txn) error {
		for _, tx := range block.Transactions {

			if !tx.IsCoinbase() {
				for _, input := range tx.Inputs {
					updatedOutputs := TxOutputs{}
					prefixedInputID := append(utxoPrefix, input.ID...)
					record, err := txn.Get(prefixedInputID)

					Handle(err)
					var outputs TxOutputs
					err = record.Value(func(val []byte) error {
						outputs = DeserializeOutputs(val)
						return nil
					})
					Handle(err)

					for outIdx, out := range outputs.Outputs {
						if input.Out != outIdx {
							updatedOutputs.Outputs = append(outputs.Outputs, out)
						}
					}

					if len(updatedOutputs.Outputs) == 0 {
						if err := txn.Delete(prefixedInputID); err != nil {
							log.Panic(err)
						}
					} else {
						if err := txn.Set(prefixedInputID, updatedOutputs.Serialize()); err != nil {
							log.Panic(err)
						}
					}
				}

			}

			newOutputs := TxOutputs{}
			newOutputs.Outputs = append(newOutputs.Outputs, tx.Outputs...)

			txID := append(utxoPrefix, tx.ID...)
			if err := txn.Set(txID, newOutputs.Serialize()); err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	Handle(err)

}

func (u UTXOSet) ReIndex() {
	db := u.Blockchain.Database
	err := u.DeleteKeysByPrefix(utxoPrefix)
	Handle(err)
	UTXO := u.Blockchain.FindUTXO()

	err = db.Update(func(txn *badger.Txn) error {
		for txID, txOutputs := range UTXO {
			key, err := hex.DecodeString(txID)
			Handle(err)
			key = append(utxoPrefix, key...)
			err = txn.Set(key, txOutputs.Serialize())
			return err
		}
		return nil
	})

	Handle(err)
}
