package common

import (
	"strings"
)

type Hexer interface {
	Hex() string
}

func ToHex(s Hexer) string {
	return strings.ToLower(s.Hex())
}
