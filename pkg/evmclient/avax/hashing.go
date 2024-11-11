package avax

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

// hasherPool holds LegacyKeccak256 hashers for rlpHash.
//
//nolint:gochecknoglobals
var hasherPool = sync.Pool{
	New: func() interface{} { return sha3.NewLegacyKeccak256() },
}

// rlpHash encodes x and hashes the encoded bytes.
//
//nolint:forcetypeassert
func rlpHash(x interface{}) (h common.Hash) {
	sha := hasherPool.Get().(crypto.KeccakState)
	defer hasherPool.Put(sha)
	sha.Reset()
	_ = rlp.Encode(sha, x)
	_, _ = sha.Read(h[:])

	return h
}
