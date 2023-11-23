package evmclient

import (
	"github.com/KyberNetwork/evmlistener/pkg/types"
	avaxtypes "github.com/ava-labs/coreth/core/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// nolint:dupl
func convertEthBlock(hash string, block *ethtypes.Block) types.Block {
	header := types.Header{
		ParentHash:       block.ParentHash().Hex(),
		UncleHash:        block.UncleHash().Hex(),
		Coinbase:         block.Coinbase().Hex(),
		StateRoot:        block.Root().Hex(),
		TransactionsRoot: block.TxHash().Hex(),
		ReceiptRoot:      block.ReceiptHash().Hex(),
		LogsBloom:        block.Bloom().Bytes(),
		Difficulty:       block.Difficulty(),
		Number:           block.Number(),
		GasLimit:         block.GasLimit(),
		GasUsed:          block.GasUsed(),
		Timestamp:        block.Time(),
		ExtraData:        block.Extra(),
		MixHash:          block.MixDigest().Hex(),
		Nonce:            block.Nonce(),
		Hash:             hash,
		BaseFeePerGas:    block.BaseFee(),
	}

	txns := make([]types.Txn, len(block.Transactions()))
	for i, tx := range block.Transactions() {
		v, r, s := tx.RawSignatureValues()

		accessList := make([]*types.AccessTuple, len(tx.AccessList()))
		for j, a := range tx.AccessList() {
			storageKeys := make([][]byte, 0, len(a.StorageKeys))
			for _, s := range a.StorageKeys {
				storageKeys = append(storageKeys, s.Bytes())
			}

			accessList[j] = &types.AccessTuple{
				Address:     a.Address.Hex(),
				StorageKeys: storageKeys,
			}
		}

		// TODO: should we skip tx if no from/to?
		// try to get sender from london signer
		from, err := ethtypes.Sender(ethtypes.NewLondonSigner(tx.ChainId()), tx)
		if err != nil {
			// try to get sender from latest signer
			from, err = ethtypes.Sender(ethtypes.LatestSignerForChainID(tx.ChainId()), tx)
			if err != nil {
				continue
			}
		}
		if tx.To() == nil {
			continue
		}

		txns[i] = types.Txn{
			To:                   tx.To().Hex(),
			Nonce:                tx.Nonce(),
			GasPrice:             tx.GasPrice(),
			GasLimit:             tx.Gas(),
			Value:                tx.Value(),
			Input:                tx.Data(),
			V:                    v.Bytes(),
			R:                    r.Bytes(),
			S:                    s.Bytes(),
			Type:                 tx.Type(),
			AccessList:           accessList,
			MaxFeePerGas:         tx.GasFeeCap(),
			MaxPriorityFeePerGas: tx.GasTipCap(),
			Hash:                 tx.Hash().Hex(),
			From:                 from.Hex(),
		}
	}

	return types.Block{
		Number:       block.Number(),
		Hash:         hash,
		Timestamp:    header.Timestamp,
		ParentHash:   header.ParentHash,
		Transactions: txns,
		Header:       header,
		ReorgedHash:  "",  // This field will be fill later
		Logs:         nil, // This field will be fill later
	}
}

// nolint:dupl
func convertAvaxBlock(hash string, block *avaxtypes.Block) types.Block {
	header := types.Header{
		ParentHash:       block.ParentHash().Hex(),
		UncleHash:        block.UncleHash().Hex(),
		Coinbase:         block.Coinbase().Hex(),
		StateRoot:        block.Root().Hex(),
		TransactionsRoot: block.TxHash().Hex(),
		ReceiptRoot:      block.ReceiptHash().Hex(),
		LogsBloom:        block.Bloom().Bytes(),
		Difficulty:       block.Difficulty(),
		Number:           block.Number(),
		GasLimit:         block.GasLimit(),
		GasUsed:          block.GasUsed(),
		Timestamp:        block.Time(),
		ExtraData:        block.Extra(),
		MixHash:          block.MixDigest().Hex(),
		Nonce:            block.Nonce(),
		Hash:             hash,
		BaseFeePerGas:    block.BaseFee(),
	}

	txns := make([]types.Txn, len(block.Transactions()))
	for i, tx := range block.Transactions() {
		v, r, s := tx.RawSignatureValues()

		accessList := make([]*types.AccessTuple, len(tx.AccessList()))
		for j, a := range tx.AccessList() {
			storageKeys := make([][]byte, 0, len(a.StorageKeys))
			for _, s := range a.StorageKeys {
				storageKeys = append(storageKeys, s.Bytes())
			}

			accessList[j] = &types.AccessTuple{
				Address:     a.Address.Hex(),
				StorageKeys: storageKeys,
			}
		}

		// TODO: should we skip tx if no from/to?
		// try to get sender from london signer
		from, err := avaxtypes.Sender(avaxtypes.NewLondonSigner(tx.ChainId()), tx)
		if err != nil {
			// try to get sender from latest signer
			from, err = avaxtypes.Sender(avaxtypes.LatestSignerForChainID(tx.ChainId()), tx)
			if err != nil {
				continue
			}
		}
		if tx.To() == nil {
			continue
		}

		txns[i] = types.Txn{
			To:                   tx.To().Hex(),
			Nonce:                tx.Nonce(),
			GasPrice:             tx.GasPrice(),
			GasLimit:             tx.Gas(),
			Value:                tx.Value(),
			Input:                tx.Data(),
			V:                    v.Bytes(),
			R:                    r.Bytes(),
			S:                    s.Bytes(),
			Type:                 tx.Type(),
			AccessList:           accessList,
			MaxFeePerGas:         tx.GasFeeCap(),
			MaxPriorityFeePerGas: tx.GasTipCap(),
			Hash:                 tx.Hash().Hex(),
			From:                 from.Hex(),
		}
	}

	return types.Block{
		Number:       block.Number(),
		Hash:         hash,
		Timestamp:    header.Timestamp,
		ParentHash:   header.ParentHash,
		Transactions: txns,
		Header:       header,
		ReorgedHash:  "",  // This field will be fill later
		Logs:         nil, // This field will be fill later
	}
}
