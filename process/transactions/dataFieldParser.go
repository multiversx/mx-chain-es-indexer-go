package transactions

import (
	"strings"

	"github.com/ElrondNetwork/elrond-go/process"
)

const (
	atSeparator = "@"
	okHexConst  = "6f6b"
)

type argumentsParserExtended struct {
	process.CallArgumentsParser
}

func newArgumentsParser(argsParser process.CallArgumentsParser) *argumentsParserExtended {
	return &argumentsParserExtended{
		CallArgumentsParser: argsParser,
	}
}

func (age *argumentsParserExtended) split(dataField string) []string {
	return strings.Split(dataField, atSeparator)
}

func (age *argumentsParserExtended) hasOKPrefix(dataField string) bool {
	splitData := age.split(dataField)
	if len(splitData) < 2 {
		return false
	}

	return splitData[1] == okHexConst
}
