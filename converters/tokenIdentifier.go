package converters

import (
	"encoding/hex"
	"fmt"
	"math/big"
)

// ComputeTokenIdentifier will compute the token identifier based on the token string and the nonce
func ComputeTokenIdentifier(token string, nonce uint64) string {
	if token == "" || nonce == 0 {
		return ""
	}

	nonceBig := big.NewInt(0).SetUint64(nonce)
	hexEncodedNonce := hex.EncodeToString(nonceBig.Bytes())
	return fmt.Sprintf("%s-%s", token, hexEncodedNonce)
}

// EncodeNonceToHex will encode provided nonce in a hex format
func EncodeNonceToHex(nonce uint64) string {
	if nonce == 0 {
		return "00"
	}

	nonceBigBytes := big.NewInt(0).SetUint64(nonce).Bytes()

	return hex.EncodeToString(nonceBigBytes)
}
