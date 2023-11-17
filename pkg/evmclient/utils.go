package evmclient

import (
	"github.com/KyberNetwork/evmlistener/protobuf/pb"
	avaxtypes "github.com/ava-labs/coreth/core/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// nolint:dupl
func ethBlockToProto(hash string, block *ethtypes.Block) *pb.Block {
	header := &pb.BlockHeader{
		ParentHash:       block.ParentHash().Bytes(),
		UncleHash:        block.UncleHash().Bytes(),
		Coinbase:         block.Coinbase().Bytes(),
		StateRoot:        block.Root().Bytes(),
		TransactionsRoot: block.TxHash().Bytes(),
		ReceiptRoot:      block.ReceiptHash().Bytes(),
		LogsBloom:        block.Bloom().Bytes(),
		Difficulty:       &pb.BigInt{Bytes: block.Difficulty().Bytes()},
		TotalDifficulty:  nil, // TODO: not found in the RPC call
		Number:           block.Number().Uint64(),
		GasLimit:         block.GasLimit(),
		GasUsed:          block.GasUsed(),
		Timestamp:        block.Time(),
		ExtraData:        block.Extra(),
		MixHash:          block.MixDigest().Bytes(),
		Nonce:            block.Nonce(),
		Hash:             []byte(hash),
		BaseFeePerGas:    &pb.BigInt{Bytes: block.BaseFee().Bytes()},
	}

	txns := make([]*pb.TransactionTrace, len(block.Transactions()))
	for i, tx := range block.Transactions() {
		v, r, s := tx.RawSignatureValues()

		accessList := make([]*pb.AccessTuple, len(tx.AccessList()))
		for j, a := range tx.AccessList() {
			storageKeys := make([][]byte, 0, len(a.StorageKeys))
			for _, s := range a.StorageKeys {
				storageKeys = append(storageKeys, s.Bytes())
			}

			accessList[j] = &pb.AccessTuple{
				Address:     a.Address.Bytes(),
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

		txns[i] = &pb.TransactionTrace{
			To:                   tx.To().Bytes(),
			Nonce:                tx.Nonce(),
			GasPrice:             &pb.BigInt{Bytes: tx.GasPrice().Bytes()},
			GasLimit:             tx.Gas(),
			Value:                &pb.BigInt{Bytes: tx.Value().Bytes()},
			Input:                tx.Data(),
			V:                    v.Bytes(),
			R:                    r.Bytes(),
			S:                    s.Bytes(),
			Type:                 pb.TransactionTrace_Type(tx.Type()),
			AccessList:           accessList,
			MaxFeePerGas:         &pb.BigInt{Bytes: tx.GasFeeCap().Bytes()},
			MaxPriorityFeePerGas: &pb.BigInt{Bytes: tx.GasTipCap().Bytes()},
			Hash:                 tx.Hash().Bytes(),
			From:                 from.Bytes(),
			TransactionIndex:     nil, // TODO: not found in the RPC call
			GasUsed:              0,   // TODO: not found in the RPC call
			Receipt:              nil, // TODO: not found in the RPC call
		}
	}

	return &pb.Block{
		Hash:              []byte(hash),
		Number:            block.NumberU64(),
		Header:            header,
		Uncles:            nil, // TODO: I don't think we need this field
		TransactionTraces: txns,
		BalanceChanges:    nil, // TODO: I don't think we need this field
		Logs:              nil, // This field will be fill later
		TraceCalls:        nil, // TODO: I don't think we need this field
		CodeChanges:       nil, // TODO: I don't think we need this field
		Ver:               0,   // TODO: I don't think we need this field
		Size:              0,   // TODO: I don't think we need this field
	}
}

// nolint:dupl
func avaxBlockToProto(hash string, block *avaxtypes.Block) *pb.Block {
	header := &pb.BlockHeader{
		ParentHash:       block.ParentHash().Bytes(),
		UncleHash:        block.UncleHash().Bytes(),
		Coinbase:         block.Coinbase().Bytes(),
		StateRoot:        block.Root().Bytes(),
		TransactionsRoot: block.TxHash().Bytes(),
		ReceiptRoot:      block.ReceiptHash().Bytes(),
		LogsBloom:        block.Bloom().Bytes(),
		Difficulty:       &pb.BigInt{Bytes: block.Difficulty().Bytes()},
		TotalDifficulty:  nil, // TODO: not found in the RPC call
		Number:           block.Number().Uint64(),
		GasLimit:         block.GasLimit(),
		GasUsed:          block.GasUsed(),
		Timestamp:        block.Time(),
		ExtraData:        block.Extra(),
		MixHash:          block.MixDigest().Bytes(),
		Nonce:            block.Nonce(),
		Hash:             []byte(hash),
		BaseFeePerGas:    &pb.BigInt{Bytes: block.BaseFee().Bytes()},
	}

	txns := make([]*pb.TransactionTrace, len(block.Transactions()))
	for i, tx := range block.Transactions() {
		v, r, s := tx.RawSignatureValues()

		accessList := make([]*pb.AccessTuple, len(tx.AccessList()))
		for j, a := range tx.AccessList() {
			storageKeys := make([][]byte, 0, len(a.StorageKeys))
			for _, s := range a.StorageKeys {
				storageKeys = append(storageKeys, s.Bytes())
			}

			accessList[j] = &pb.AccessTuple{
				Address:     a.Address.Bytes(),
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

		txns[i] = &pb.TransactionTrace{
			To:                   tx.To().Bytes(),
			Nonce:                tx.Nonce(),
			GasPrice:             &pb.BigInt{Bytes: tx.GasPrice().Bytes()},
			GasLimit:             tx.Gas(),
			Value:                &pb.BigInt{Bytes: tx.Value().Bytes()},
			Input:                tx.Data(),
			V:                    v.Bytes(),
			R:                    r.Bytes(),
			S:                    s.Bytes(),
			Type:                 pb.TransactionTrace_Type(tx.Type()),
			AccessList:           accessList,
			MaxFeePerGas:         &pb.BigInt{Bytes: tx.GasFeeCap().Bytes()},
			MaxPriorityFeePerGas: &pb.BigInt{Bytes: tx.GasTipCap().Bytes()},
			Hash:                 tx.Hash().Bytes(),
			From:                 from.Bytes(),
			TransactionIndex:     nil, // TODO: not found in the RPC call
			GasUsed:              0,   // TODO: not found in the RPC call
			Receipt:              nil, // TODO: not found in the RPC call
		}
	}

	return &pb.Block{
		Hash:              []byte(hash),
		Number:            block.NumberU64(),
		Header:            header,
		Uncles:            nil, // TODO: I don't think we need this field
		TransactionTraces: txns,
		BalanceChanges:    nil, // TODO: I don't think we need this field
		Logs:              nil, // This field will be fill later
		TraceCalls:        nil, // TODO: I don't think we need this field
		CodeChanges:       nil, // TODO: I don't think we need this field
		Ver:               0,   // TODO: I don't think we need this field
		Size:              0,   // TODO: I don't think we need this field
	}
}
