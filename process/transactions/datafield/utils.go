package datafield

import (
	"bytes"
	"fmt"
	"math/big"
	"unicode"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

const (
	esdtIdentifierSeparator  = "-"
	esdtRandomSequenceLength = 6
)

func extractTokenIdentifierAndNonce(arg []byte) ([]byte, uint64) {
	argsSplit := bytes.Split(arg, []byte(esdtIdentifierSeparator))
	if len(argsSplit) < 2 {
		return arg, 0
	}

	if len(argsSplit[1]) <= esdtRandomSequenceLength {
		return arg, 0
	}

	identifier := []byte(fmt.Sprintf("%s-%s", argsSplit[0], argsSplit[1][:esdtRandomSequenceLength]))
	nonce := big.NewInt(0).SetBytes(argsSplit[1][esdtRandomSequenceLength:])

	return identifier, nonce.Uint64()
}

func isEmptyAddr(pubKeyConverter core.PubkeyConverter, receiver []byte) bool {
	emptyAddr := make([]byte, pubKeyConverter.Len())

	return bytes.Equal(receiver, emptyAddr)
}

func isASCIIString(input string) bool {
	for i := 0; i < len(input); i++ {
		if input[i] > unicode.MaxASCII {
			return false
		}
	}

	return true
}
