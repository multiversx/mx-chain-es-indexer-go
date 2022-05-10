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

func getAllBuiltInFunctions() []string {
	return []string{
		core.BuiltInFunctionClaimDeveloperRewards,
		core.BuiltInFunctionChangeOwnerAddress,
		core.BuiltInFunctionSetUserName,
		core.BuiltInFunctionSaveKeyValue,
		core.BuiltInFunctionESDTTransfer,
		core.BuiltInFunctionESDTBurn,
		core.BuiltInFunctionESDTFreeze,
		core.BuiltInFunctionESDTUnFreeze,
		core.BuiltInFunctionESDTWipe,
		core.BuiltInFunctionESDTPause,
		core.BuiltInFunctionESDTUnPause,
		core.BuiltInFunctionSetESDTRole,
		core.BuiltInFunctionUnSetESDTRole,
		core.BuiltInFunctionESDTSetLimitedTransfer,
		core.BuiltInFunctionESDTUnSetLimitedTransfer,
		core.BuiltInFunctionESDTLocalMint,
		core.BuiltInFunctionESDTLocalBurn,
		core.BuiltInFunctionESDTNFTTransfer,
		core.BuiltInFunctionESDTNFTCreate,
		core.BuiltInFunctionESDTNFTAddQuantity,
		core.BuiltInFunctionESDTNFTCreateRoleTransfer,
		core.BuiltInFunctionESDTNFTBurn,
		core.BuiltInFunctionESDTNFTAddURI,
		core.BuiltInFunctionESDTNFTUpdateAttributes,
		core.BuiltInFunctionMultiESDTNFTTransfer,
		core.ESDTRoleLocalMint,
		core.ESDTRoleLocalBurn,
		core.ESDTRoleNFTCreate,
		core.ESDTRoleNFTCreateMultiShard,
		core.ESDTRoleNFTAddQuantity,
		core.ESDTRoleNFTBurn,
		core.ESDTRoleNFTAddURI,
		core.ESDTRoleNFTUpdateAttributes,
		core.ESDTRoleTransfer,
	}
}

func isBuiltInFunction(builtInFunctionsList []string, function string) bool {
	for _, builtInFunction := range builtInFunctionsList {
		if builtInFunction == function {
			return true
		}
	}

	return false
}

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
