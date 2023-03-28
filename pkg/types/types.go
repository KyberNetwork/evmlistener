package types

import (
	"math/big"
)

// Log contains log information.
type Log struct {
	Address     string   `json:"address"`
	Topics      []string `json:"topics"`
	Data        []byte   `json:"data"`
	BlockNumber uint64   `json:"blockNumber"`
	TxHash      string   `json:"transactionHash"`
	TxIndex     uint     `json:"transactionIndex"`
	BlockHash   string   `json:"blockHash"`
	Index       uint     `json:"logIndex"`
	Removed     bool     `json:"removed"`
}

// Header contains block header information.
type Header struct {
	Hash       string   `json:"hash"`
	ParentHash string   `json:"parentHash"`
	Number     *big.Int `json:"number"`
	Time       uint64   `json:"timestamp"`
}

// Block contains information of block.
type Block struct {
	Number      *big.Int `json:"number"`
	Hash        string   `json:"hash"`
	Timestamp   uint64   `json:"timestamp"`
	ParentHash  string   `json:"parentHash"`
	ReorgedHash string   `json:"reorgedHash"`
	Logs        []Log    `json:"logs"`
}

// Message ...
type Message struct {
	RevertedBlocks []Block `json:"revertedBlocks"`
	NewBlocks      []Block `json:"newBlocks"`
}
