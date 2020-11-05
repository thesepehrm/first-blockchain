package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	"gitlab.com/thesepehrm/first-blockchain/wallet"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func (tx Transaction) String() string {
	lines := []string{}

	lines = append(lines, fmt.Sprintf("—— Transaction %x:", tx.ID))
	for inputId, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("\tInput %d", inputId))
		lines = append(lines, fmt.Sprintf("\t\tTXID: %x", input.ID))
		lines = append(lines, fmt.Sprintf("\t\tOut: %d", input.Out))
		lines = append(lines, fmt.Sprintf("\t\tSignature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("\t\tPubKey: %x", input.PubKey))
	}

	for outputId, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("\tOutput %d", outputId))
		lines = append(lines, fmt.Sprintf("\t\tValue: %d", output.Value))
		lines = append(lines, fmt.Sprintf("\t\tPubKeyHash: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")

}

func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		Handle(err)
		data = fmt.Sprintf("%x", randData)
	}

	txin := TxInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTxOutput(10, to)

	tx := Transaction{nil, []TxInput{txin}, []TxOutput{*txout}}

	tx.ID = tx.Hash()

	return &tx
}

func NewTransaction(from, to string, amount int, UTXO *UTXOSet) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	wallets, err := wallet.CreateWallets()
	Handle(err)
	w := wallets.GetWallet(from)
	pubKeyHash := wallet.PublicKeyHash(w.PublicKey)

	acc, validOutputs := UTXO.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("Error: not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)

		for _, out := range outs {
			input := TxInput{txID, out, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, *NewTxOutput(amount, to))

	if acc > amount {
		outputs = append(outputs, *NewTxOutput(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	UTXO.Blockchain.SignTransaction(&tx, w.PrivateKey)

	return &tx
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

func (tx Transaction) Serialize() []byte {
	var content bytes.Buffer

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(tx)
	Handle(err)

	return content.Bytes()
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	tempTx := *tx
	tempTx.ID = []byte{}

	hash = sha256.Sum256(tempTx.Serialize())

	return hash[:]
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, txInput := range tx.Inputs {
		inputs = append(inputs, txInput)
	}

	for _, txOutput := range tx.Outputs {
		outputs = append(outputs, txOutput)
	}

	txCopy := Transaction{[]byte{}, inputs, outputs}
	return txCopy

}

func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTxs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, in := range prevTxs {
		if prevTxs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("Error: Previous transaction does not exist")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inId, in := range txCopy.Inputs {
		prevTx := prevTxs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTx.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)
		Handle(err)
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Inputs[inId].Signature = signature
	}
}

func (tx *Transaction) Verify(prevTxs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, in := range prevTxs {
		if prevTxs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("Error: Previous transaction does not exist")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inId, in := range txCopy.Inputs {
		prevTx := prevTxs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTx.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		// first half of the signature is r and the second half is s
		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen / 2)])
		s.SetBytes(in.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}

		pubLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:(pubLen / 2)])
		y.SetBytes(in.PubKey[(pubLen / 2):])

		rawPublicKey := ecdsa.PublicKey{curve, &x, &y}

		if ecdsa.Verify(&rawPublicKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}

	return true
}
