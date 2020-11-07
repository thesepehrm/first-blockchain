package network

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"runtime"
	"syscall"

	"gitlab.com/thesepehrm/first-blockchain/blockchain"
	"gopkg.in/vrecan/death.v3"
)

const (
	protocol      = "tcp"
	version       = 1
	commandLength = 12
)

type nodes []string

var (
	nodeAddress     string
	minerAddress    string
	knownNodes      = nodes{"localhost:3000"}
	blocksInTransit = [][]byte{}
	memoryPool      = make(map[string]blockchain.Transaction)
)

type Addr struct {
	AddrList []string
}

type Block struct {
	AddrFrom string
	Block    []byte
}

type GetBlocks struct {
	AddrFrom string
}

type GetData struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type Inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type Tx struct {
	AddrFrom    string
	Transaction []byte
}

type Version struct {
	AddrFrom  string
	Version   int
	BestHight int
}

func (array nodes) InArray(key string) bool {
	for _, element := range array {
		if element == key {
			return true
		}
	}

	return false
}

func CloseDB(chain *blockchain.BlockChain) {
	d := death.NewDeath(syscall.SIGINT, syscall.SIGKILL, os.Interrupt)

	d.WaitForDeathWithFunc(func() {
		defer os.Exit(1)
		defer runtime.Goexit()
		chain.Database.Close()
	})
}

func HandleConnection(conn net.Conn, chain *blockchain.BlockChain) {
	req, err := ioutil.ReadAll(conn)
	Handle(err)

	command := BytesToCmd(req[:commandLength])
	fmt.Printf("Requested command: %s", command)

	switch command {
	case "block":
		HandleBlock(req, chain)
	case "addr":
		HandleAddr(req)
	case "getblocks":
		HandleGetBlocks(req, chain)
	case "getdata":
		HandleGetData(req, chain)
	case "inv":
		HandleInv(req, chain)
	case "tx":
		HandleTx(req, chain)
	case "version":
		HandleVersion(req, chain)

	default:
		fmt.Println("Unknown command")
	}

}

func Start(nodeID, minerNodeAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	minerAddress = minerNodeAddress

	listener, err := net.Listen(protocol, nodeAddress)
	Handle(err)
	defer listener.Close()
	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()
	go CloseDB(chain)

	if nodeAddress != knownNodes[0] {
		SendVersion(knownNodes[0], chain)
	} else {
		for {
			conn, err := listener.Accept()
			Handle(err)
			go HandleConnection(conn, chain)
		}
	}
}

func SendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)

		for _, node := range knownNodes {
			if node != addr {
				knownNodes = append(knownNodes, node)
			}
		}

		return
	}

	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	Handle(err)

}

func BuildAndSendData(addr, command string, data interface{}) {
	payload := GobEncode(data)
	req := append(CmdToBytes(command), payload...)
	SendData(addr, req)
}

func SendAddr(addr string) {

	data := Addr{append(knownNodes, addr)}
	BuildAndSendData(addr, "addr", data)
}

func SendBlock(addr string, block *blockchain.Block) {
	data := Block{nodeAddress, block.Serialize()}
	BuildAndSendData(addr, "block", data)
}

func SendInv(addr string, kind string, items [][]byte) {
	data := Inv{nodeAddress, kind, items}
	BuildAndSendData(addr, "inv", data)
}

func SendTX(addr string, tx *blockchain.Transaction) {
	data := Tx{nodeAddress, tx.Serialize()}
	BuildAndSendData(addr, "tx", data)
}

func SendVersion(addr string, chain *blockchain.BlockChain) {
	bestHeight := chain.GetBestHeight()
	data := Version{nodeAddress, version, bestHeight}
	BuildAndSendData(addr, "version", data)
}

func SendGetBlocks(addr string) {
	data := GetBlocks{nodeAddress}
	BuildAndSendData(addr, "getblocks", data)
}

func RequestBlocks() {

	for _, node := range knownNodes {
		SendGetBlocks(node)
	}
}

func SendGetData(addr, kind string, id []byte) {
	data := GetData{nodeAddress, kind, id}
	BuildAndSendData(addr, "getdata", data)
}

func HandleAddr(req []byte) {
	var buffer bytes.Buffer
	var payload Addr

	buffer.Write(req[commandLength:])

	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(payload)
	Handle(err)

	knownNodes = append(knownNodes, payload.AddrList...)

	fmt.Printf("There are %d known nodes in the list.", len(knownNodes))
	RequestBlocks()
}

func HandleBlock(req []byte, chain *blockchain.BlockChain) {
	var buffer bytes.Buffer
	var payload Block

	buffer.Write(req[commandLength:])

	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(payload)
	Handle(err)

	block := blockchain.Deserialize(payload.Block)

	chain.AddBlock(block)
	fmt.Printf("New Block Received and added to chain: %x", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		SendGetData(payload.AddrFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]

	} else {
		UTXOSet := blockchain.UTXOSet{chain}
		UTXOSet.ReIndex()
	}
}

func HandleGetBlocks(req []byte, chain *blockchain.BlockChain) {
	var buffer bytes.Buffer
	var payload GetBlocks

	buffer.Write(req[commandLength:])

	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(payload)
	Handle(err)

	blocks := chain.GetBlockHashes()
	SendInv(payload.AddrFrom, "block", blocks)
}

func HandleGetData(req []byte, chain *blockchain.BlockChain) {
	var buffer bytes.Buffer
	var payload GetData

	buffer.Write(req[commandLength:])

	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(payload)
	Handle(err)

	switch payload.Type {
	case "block":
		block, err := chain.GetBlock([]byte(payload.ID))
		Handle(err)

		SendBlock(payload.AddrFrom, &block)
	case "tx":
		txID := hex.EncodeToString(payload.ID)
		tx := memoryPool[txID]

		SendTX(payload.AddrFrom, &tx)
	}

}

func HandleVersion(req []byte, chain *blockchain.BlockChain) {
	var buffer bytes.Buffer
	var payload Version

	buffer.Write(req[commandLength:])

	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(payload)
	Handle(err)

	bestHeight := chain.GetBestHeight()
	sentHeight := payload.BestHight

	if bestHeight > sentHeight {
		SendVersion(payload.AddrFrom, chain)
	} else if bestHeight < sentHeight {
		SendGetBlocks(payload.AddrFrom)
	}

	if !knownNodes.InArray(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}

}

func HandleTx(req []byte, chain *blockchain.BlockChain) {
	var buffer bytes.Buffer
	var payload Tx

	buffer.Write(req[commandLength:])

	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(payload)
	Handle(err)

	txData := payload.Transaction
	tx := blockchain.DeserializeTransaction(txData)
	memoryPool[hex.EncodeToString(tx.ID)] = tx

	if nodeAddress == knownNodes[0] { // Main full node
		for _, node := range knownNodes {
			if node != nodeAddress && node != payload.AddrFrom {
				SendInv(node, "tx", []byte(tx.ID))
			}

		}
	} else {
		if len(memoryPool) > 2 && len(minerAddress) > 0 {
			MineTx(chain)
		}
	}

}

func MineTx(chain *blockchain.BlockChain) {
	var txs []*blockchain.Transaction

	for id := range memoryPool {
		fmt.Printf("tx: %s\n", memoryPool[id].ID)
		tx := memoryPool[id]
		if chain.VerifyTransaction(&tx) {
			txs = append(txs, &tx)
		}
	}

	if len(txs) == 0 {
		fmt.Println("All Transactions are invalid")
		return
	}

	cbTx := blockchain.CoinbaseTx(minerAddress, "")
	txs = append(txs, cbTx)

	newBlock := chain.MineBlock(txs)
	UTXOSet := blockchain.UTXOSet{chain}
	UTXOSet.ReIndex()

	fmt.Println("New Block mined")

	for _, tx := range txs {
		txID := hex.EncodeToString(tx.ID)
		delete(memoryPool, txID)
	}

	for _, node := range knownNodes {
		if node != nodeAddress {
			SendInv(node, "block", [][]byte{newBlock.Hash})
		}
	}

	if len(memoryPool) > 0 {
		MineTx(chain)
	}
}

func HandleInv(request []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload Inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	Handle(err)

	fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		SendGetData(payload.AddrFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, block := range blocksInTransit {
			if bytes.Compare(block, blockHash) != 0 {
				newInTransit = append(newInTransit, block)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if memoryPool[hex.EncodeToString(txID)].ID == nil {
			SendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}
