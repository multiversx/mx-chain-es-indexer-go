package transactions

import (
	"strings"

	"github.com/ElrondNetwork/elrond-vm-common/parsers"
)

const (
	atSeparator  = "@"
	okHexConst   = "6f6b"
	zeroHexConst = "00"
)

type parser interface {
	ParseData(data string) (string, [][]byte, error)
}

type argumentsParserExtended struct {
	parser
}

func newArgumentsParser() *argumentsParserExtended {
	return &argumentsParserExtended{
		parser: parsers.NewCallArgsParser(),
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

func (age *argumentsParserExtended) hasZeroPrefix(dataField string) bool {
	splitData := age.split(dataField)
	if len(splitData) < 2 {
		return false
	}

	return splitData[1] == zeroHexConst
}
