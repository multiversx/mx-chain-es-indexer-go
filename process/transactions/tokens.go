package transactions

import (
	"encoding/hex"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
)

const (
	issueFungibleESDTFunc     = "issue"
	issueSemiFungibleESDTFunc = "issueSemiFungible"
	issueNonFungibleESDTFunc  = "issueNonFungible"
)

type tokensProcessor struct {
	issueMethods           map[string]struct{}
	argumentParserExtended *argumentsParserExtended
	selfShardID            uint32
	esdtSCAddress          string
}

func newTokensProcessor(selfShardID uint32, pubKeyConverter core.PubkeyConverter) *tokensProcessor {
	argsParser := parsers.NewCallArgsParser()

	return &tokensProcessor{
		selfShardID: selfShardID,
		issueMethods: map[string]struct{}{
			issueFungibleESDTFunc:     {},
			issueSemiFungibleESDTFunc: {},
			issueNonFungibleESDTFunc:  {},
		},
		argumentParserExtended: newArgumentsParser(argsParser),
		esdtSCAddress:          pubKeyConverter.Encode(vm.ESDTSCAddress),
	}
}

func (tp *tokensProcessor) searchForTokenIssueTransactions(txs []*data.Transaction, timestamp uint64) []*data.TokenInfo {
	if tp.selfShardID != core.MetachainShardId {
		return []*data.TokenInfo{}
	}

	tokensData := make([]*data.TokenInfo, 0)
	for _, tx := range txs {
		funcName, okSCR, ok := tp.isIssueTxSuccess(tx.Data, tx.Sender, tx.Nonce, tx.SmartContractResults)
		if !ok {
			continue
		}

		tokenInfo, ok := tp.extractTokenInfo(tx.Data, okSCR)
		if !ok {
			continue
		}

		if funcName == issueFungibleESDTFunc {
			tokenInfo.Token = tp.searchTokenIdentifierFungibleESDT(tx.Sender, tx.SmartContractResults)
		}

		tokenInfo.Timestamp = time.Duration(timestamp)
		tokensData = append(tokensData, tokenInfo)
	}

	return tokensData
}

func (tp *tokensProcessor) searchForTokenIssueScrs(scrs []*data.ScResult, timestamp uint64) []*data.TokenInfo {
	if tp.selfShardID != core.MetachainShardId {
		return []*data.TokenInfo{}
	}

	tokensData := make([]*data.TokenInfo, 0)
	for _, scr := range scrs {
		funcName, resScr, ok := tp.isIssueSCRSuccess(scr.Data, scr.Sender, scrs)
		if !ok {
			continue
		}

		tokenInfo, ok := tp.extractTokenInfo(scr.Data, resScr)
		if !ok {
			continue
		}

		tokenInfo.Timestamp = time.Duration(timestamp)

		if funcName != issueFungibleESDTFunc {
			tokensData = append(tokensData, tokenInfo)
			continue
		}

		identifier := tp.searchTokenIdentifierFungibleESDT(scr.Sender, scrs)
		if identifier == "" {
			continue
		}
		tokenInfo.Issuer = scr.Sender
		tokenInfo.Token = identifier

		tokensData = append(tokensData, tokenInfo)
	}

	return tokensData
}

func (tp *tokensProcessor) searchTokenIdentifierFungibleESDT(sender string, scrs []*data.ScResult) string {
	for _, scr := range scrs {
		isReceiverAndSenderOk := scr.Receiver == sender && scr.Sender == tp.esdtSCAddress
		if !isReceiverAndSenderOk {
			continue
		}
		if scr.Nonce != 0 {
			continue
		}

		funcName, args, err := tp.argumentParserExtended.ParseData(string(scr.Data))
		if err != nil {
			continue
		}

		if funcName != core.BuiltInFunctionESDTTransfer || len(args) < 2 {
			continue
		}

		return string(args[0])
	}

	return ""
}

func (tp *tokensProcessor) extractTokenInfo(dataField []byte, scr *data.ScResult) (*data.TokenInfo, bool) {
	funcName, args, err := tp.argumentParserExtended.ParseData(string(dataField))
	if err != nil {
		return nil, false
	}

	if len(args) < 2 {
		return nil, false
	}

	tokenInfo := &data.TokenInfo{
		Name:   string(args[0]),
		Ticker: string(args[1]),
		Type:   computeTokenTypeByIssueFunction(funcName),
	}

	if scr == nil {
		return tokenInfo, true
	}

	tokenInfo.Issuer = scr.Receiver

	split := tp.argumentParserExtended.split(string(scr.Data))
	if len(split) < 3 {
		return tokenInfo, true
	}

	tokenIdentifier, err := hex.DecodeString(split[2])
	if err != nil {
		return nil, false
	}

	tokenInfo.Token = string(tokenIdentifier)

	return tokenInfo, true
}

func (tp *tokensProcessor) isIssueTxSuccess(dataField []byte, sender string, nonce uint64, scrs []*data.ScResult) (string, *data.ScResult, bool) {
	funcName, ok := tp.isFuncNameOK(dataField)
	if !ok {
		return "", nil, false
	}

	for _, scr := range scrs {
		areReceiverAndSenderOk := scr.Receiver == sender && scr.Sender == tp.esdtSCAddress
		if !areReceiverAndSenderOk {
			continue
		}

		isNonceOk := scr.Nonce == nonce+1
		if !isNonceOk {
			continue
		}

		if tp.argumentParserExtended.hasOKPrefix(string(scr.Data)) {
			return funcName, scr, true
		}
	}

	return "", nil, false
}

func (tp *tokensProcessor) isFuncNameOK(dataField []byte) (string, bool) {
	funcName, _, err := tp.argumentParserExtended.ParseData(string(dataField))
	if err != nil {
		return "", false
	}

	_, ok := tp.issueMethods[funcName]
	if !ok {
		return "", false
	}

	return funcName, true
}

func (tp *tokensProcessor) isIssueSCRSuccess(dataField []byte, sender string, scrs []*data.ScResult) (string, *data.ScResult, bool) {
	funcName, ok := tp.isFuncNameOK(dataField)
	if !ok {
		return "", nil, false
	}
	if funcName == issueFungibleESDTFunc {
		return funcName, nil, true
	}

	for _, scr := range scrs {
		areReceiverAndSenderOk := scr.Receiver == sender && scr.Sender == tp.esdtSCAddress
		if !areReceiverAndSenderOk {
			continue
		}

		if tp.argumentParserExtended.hasZeroPrefix(string(scr.Data)) {
			return funcName, scr, true
		}
	}

	return "", nil, false
}

func computeTokenTypeByIssueFunction(issueFuncName string) string {
	switch issueFuncName {
	case issueFungibleESDTFunc:
		return core.FungibleESDT
	case issueSemiFungibleESDTFunc:
		return core.SemiFungibleESDT
	case issueNonFungibleESDTFunc:
		return core.NonFungibleESDT
	default:
		return ""
	}
}
