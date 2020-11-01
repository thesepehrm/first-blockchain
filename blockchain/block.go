package blockchain

// BlockChain is the structure for a blockchain
type BlockChain struct {
	Blocks []*Block
}

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

func (chain *BlockChain) AddBlock(data string) {
	prevBlock := chain.Blocks[len(chain.Blocks)-1]
	newBlock := createBlock(data, prevBlock.Hash)
	chain.Blocks = append(chain.Blocks, newBlock)
}

// Genesis creates the genesis block
func Genesis() *Block {
	return createBlock("Genesis", []byte{})
}

// InitBlockChain makes a new blockchain
func InitBlockChain() *BlockChain {
	return &BlockChain{[]*Block{Genesis()}}
}
