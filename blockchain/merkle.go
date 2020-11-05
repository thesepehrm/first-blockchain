package blockchain

import (
	"crypto/sha256"
	"log"
)

type MerkleTree struct {
	Root *MerkleNode
}

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

func NewMerkleNode(left *MerkleNode, right *MerkleNode, data []byte) *MerkleNode {
	node := new(MerkleNode)
	var hash [32]byte

	if left == nil && right == nil {
		hash = sha256.Sum256(data)
	} else {
		if left == nil || right == nil {
			log.Panic("All nodes must either have two children or nothing")
		}
		prevHashes := append(left.Data, right.Data...)
		hash = sha256.Sum256(prevHashes)
	}

	node.Left = left
	node.Right = right
	node.Data = hash[:]

	return node
}

func NewMerkleTree(data [][]byte) *MerkleTree {
	merkeTree := new(MerkleTree)

	var merkleRow []*MerkleNode

	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, leaf := range data {
		merkleRow = append(merkleRow, NewMerkleNode(nil, nil, leaf))
	}

	for len(merkleRow) > 1 {
		if len(merkleRow)%2 != 0 {
			merkleRow = append(merkleRow, merkleRow[len(merkleRow)-1])
		}

		var tempRow []*MerkleNode
		for nodeID := 0; nodeID < len(merkleRow); nodeID += 2 {
			tempRow = append(tempRow, NewMerkleNode(merkleRow[nodeID], merkleRow[nodeID+1], []byte{}))
		}
		merkleRow = tempRow
	}
	merkeTree.Root = merkleRow[0]
	return merkeTree
}
