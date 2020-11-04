package blockchain

import (
	"bytes"
	"encoding/gob"

	"gitlab.com/thesepehrm/first-blockchain/wallet"
)

type TxInput struct {
	ID        []byte
	Out       int // Output index
	Signature []byte
	PubKey    []byte
}

type TxOutputs struct {
	Outputs []TxOutput
}

type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

func NewTxOutput(value int, address string) *TxOutput {
	txo := new(TxOutput)
	txo.Value = value
	txo.Lock([]byte(address))

	return txo
}

func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockHash := wallet.PublicKeyHash(in.PubKey)
	return bytes.Compare(lockHash, pubKeyHash) == 0
}

func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := wallet.DecodeBase58(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TxOutput) isLockedWith(pubkeyHash []byte) bool {
	return bytes.Compare(pubkeyHash, out.PubKeyHash) == 0
}

func (outs *TxOutputs) Serialize() []byte {
	var content bytes.Buffer

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(outs)
	Handle(err)

	return content.Bytes()
}

func DeserializeOutputs(data []byte) TxOutputs {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	outputs := TxOutputs{}
	err := decoder.Decode(&outputs)
	Handle(err)
	return outputs

}
