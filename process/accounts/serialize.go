package accounts

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

// SerializeNFTCreateInfo will serialize the provided nft create information in a way that Elastic Search expects a bulk request
func (ap *accountsProcessor) SerializeNFTCreateInfo(tokensInfo []*data.TokenInfo) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for _, tokenData := range tokensInfo {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, tokenData.Identifier, "\n"))
		serializedData, errMarshal := json.Marshal(tokenData)
		if errMarshal != nil {
			return nil, errMarshal
		}

		err := buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

// SerializeAccounts will serialize the provided accounts in a way that Elastic Search expects a bulk request
func (ap *accountsProcessor) SerializeAccounts(accounts map[string]*data.AccountInfo) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for _, acc := range accounts {
		meta, serializedData, err := prepareSerializedAccount(acc, false)
		if err != nil {
			return nil, err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

func (ap *accountsProcessor) SerializeAccountsESDT(
	accounts map[string]*data.AccountInfo,
	updateNFTData []*data.UpdateNFTData,
) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for _, acc := range accounts {
		meta, serializedData, err := prepareSerializedAccount(acc, true)
		if err != nil {
			return nil, err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	err := converters.PrepareNFTUpdateData(buffSlice, updateNFTData, true)
	if err != nil {
		return nil, err
	}

	return buffSlice.Buffers(), nil
}

func prepareSerializedAccount(acc *data.AccountInfo, isESDT bool) ([]byte, []byte, error) {
	if (acc.Balance == "0" || acc.Balance == "") && isESDT {
		meta := prepareDeleteAccountInfo(acc, isESDT)
		return meta, nil, nil
	}

	return prepareSerializedAccountInfo(acc, isESDT)
}

func prepareDeleteAccountInfo(acct *data.AccountInfo, isESDT bool) []byte {
	id := acct.Address
	if isESDT {
		hexEncodedNonce := converters.EncodeNonceToHex(acct.TokenNonce)
		id += fmt.Sprintf("-%s-%s", acct.TokenName, hexEncodedNonce)
	}

	meta := []byte(fmt.Sprintf(`{ "delete" : { "_id" : "%s" } }%s`, id, "\n"))

	return meta
}

func prepareSerializedAccountInfo(
	account *data.AccountInfo,
	isESDTAccount bool,
) ([]byte, []byte, error) {
	id := account.Address
	if isESDTAccount {
		hexEncodedNonce := converters.EncodeNonceToHex(account.TokenNonce)
		id += fmt.Sprintf("-%s-%s", account.TokenName, hexEncodedNonce)
	}

	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, id, "\n"))
	serializedData, err := json.Marshal(account)
	if err != nil {
		return nil, nil, err
	}

	return meta, serializedData, nil
}

// SerializeAccountsHistory will serialize accounts history in a way that Elastic Search expects a bulk request
func (ap *accountsProcessor) SerializeAccountsHistory(accounts map[string]*data.AccountBalanceHistory) ([]*bytes.Buffer, error) {
	var err error

	buffSlice := data.NewBufferSlice()
	for _, acc := range accounts {
		meta, serializedData, errPrepareAcc := prepareSerializedAccountBalanceHistory(acc)
		if errPrepareAcc != nil {
			return nil, err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

func prepareSerializedAccountBalanceHistory(
	account *data.AccountBalanceHistory,
) ([]byte, []byte, error) {
	id := account.Address

	isESDT := account.Token != ""
	if isESDT {
		hexEncodedNonce := converters.EncodeNonceToHex(account.TokenNonce)
		id += fmt.Sprintf("-%s-%s", account.Token, hexEncodedNonce)
	}

	id += fmt.Sprintf("-%d", account.Timestamp)
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, id, "\n"))

	serializedData, err := json.Marshal(account)
	if err != nil {
		return nil, nil, err
	}

	return meta, serializedData, nil
}
