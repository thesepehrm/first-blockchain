package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
)

const walletDBFile = "./tmp/wallets_%s.data"

type Wallets struct {
	Wallets map[string]*Wallet
}

func (ws *Wallets) SaveFile(nodeID string) {
	var content bytes.Buffer

	walletPath := fmt.Sprintf(walletDBFile, nodeID)

	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	Handle(err)

	err = ioutil.WriteFile(walletPath, content.Bytes(), 0644)
	Handle(err)

}

func (ws *Wallets) LoadFile(nodeID string) error {

	walletPath := fmt.Sprintf(walletDBFile, nodeID)

	if _, err := os.Stat(walletPath); os.IsNotExist(err) {
		return err
	}

	var fileContent []byte
	var wallets Wallets

	fileContent, err := ioutil.ReadFile(walletPath)
	Handle(err)

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	Handle(err)

	ws.Wallets = wallets.Wallets

	return nil
}

func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

func (ws *Wallets) GetAllAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

func (ws Wallets) AddWallet() string {
	wallet := MakeWallet()
	address := fmt.Sprintf("%s", wallet.Address())

	ws.Wallets[address] = wallet

	return address
}

func CreateWallets(nodeID string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFile(nodeID)
	return &wallets, err
}
