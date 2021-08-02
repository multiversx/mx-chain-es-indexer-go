package accounts

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/stretchr/testify/require"
)

func TestSerializeNFTCreateInfo(t *testing.T) {
	t.Parallel()

	nftsCreateInfo := []*data.TokenInfo{
		{
			Token:      "my-token-0001",
			Identifier: "my-token-001-0f",
			Data: &data.TokenMetaData{
				Creator: "010102",
			},
			Type: core.NonFungibleESDT,
		},
	}

	res, err := (&accountsProcessor{}).SerializeNFTCreateInfo(nftsCreateInfo)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "my-token-001-0f" } }
{"identifier":"my-token-001-0f","token":"my-token-0001","type":"NonFungibleESDT","data":{"creator":"010102","nonEmptyURIs":false}}
`
	require.Equal(t, expectedRes, res[0].String())
}

func TestSerializeAccounts(t *testing.T) {
	t.Parallel()

	accs := map[string]*data.AccountInfo{
		"addr1": {
			Address:                  "addr1",
			Nonce:                    1,
			Balance:                  "50",
			BalanceNum:               0.1,
			TotalBalanceWithStake:    "50",
			TotalBalanceWithStakeNum: 0.1,
			IsSmartContract:          true,
			IsSender:                 true,
		},
	}

	res, err := (&accountsProcessor{}).SerializeAccounts(accs, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "addr1" } }
{"address":"addr1","nonce":1,"balance":"50","balanceNum":0.1,"totalBalanceWithStake":"50","totalBalanceWithStakeNum":0.1}
`
	require.Equal(t, expectedRes, res[0].String())
}

func TestSerializeAccountsESDTNonceZero(t *testing.T) {
	t.Parallel()

	accs := map[string]*data.AccountInfo{
		"addr1": {
			Address:    "addr1",
			Nonce:      1,
			TokenName:  "token-abcd",
			Properties: "000",
			TokenNonce: 0,
			Balance:    "10000000000000",
			BalanceNum: 1,
		},
	}

	res, err := (&accountsProcessor{}).SerializeAccounts(accs, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "addr1-token-abcd-00" } }
{"address":"addr1","nonce":1,"balance":"10000000000000","balanceNum":1,"token":"token-abcd","properties":"000"}
`
	require.Equal(t, expectedRes, res[0].String())
}

func TestSerializeAccountsESDT(t *testing.T) {
	t.Parallel()

	accs := map[string]*data.AccountInfo{
		"addr1": {
			Address:    "addr1",
			Nonce:      1,
			TokenName:  "token-0001",
			Properties: "000",
			TokenNonce: 5,
			Balance:    "10000000000000",
			BalanceNum: 1,
		},
	}

	res, err := (&accountsProcessor{}).SerializeAccounts(accs, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "addr1-token-0001-05" } }
{"address":"addr1","nonce":1,"balance":"10000000000000","balanceNum":1,"token":"token-0001","tokenNonce":5,"properties":"000"}
`
	require.Equal(t, expectedRes, res[0].String())
}

func TestSerializeAccountsNFTWithMedaData(t *testing.T) {
	t.Parallel()

	accs := map[string]*data.AccountInfo{
		"addr1": {
			Address:         "addr1",
			Nonce:           1,
			TokenName:       "token-0001",
			Properties:      "000",
			TokenNonce:      22,
			Balance:         "10000000000000",
			BalanceNum:      1,
			TokenIdentifier: "token-0001-5",
			Data: &data.TokenMetaData{
				Name:      "nft",
				Creator:   "010101",
				Royalties: 1,
				Hash:      []byte("hash"),
				URIs: [][]byte{
					[]byte("uri"),
				},
				Attributes:   []byte("tags:test,free,fun;description:This is a test description for an awesome nft"),
				Tags:         converters.ExtractTagsFromAttributes([]byte("tags:test,free,fun;description:This is a test description for an awesome nft")),
				MetaData:     "metadata-test",
				NonEmptyURIs: true,
			},
		},
	}

	res, err := (&accountsProcessor{}).SerializeAccounts(accs, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "addr1-token-0001-16" } }
{"address":"addr1","nonce":1,"balance":"10000000000000","balanceNum":1,"token":"token-0001","identifier":"token-0001-5","tokenNonce":22,"properties":"000","data":{"name":"nft","creator":"010101","royalties":1,"hash":"aGFzaA==","uris":["dXJp"],"tags":["test","free","fun"],"attributes":"dGFnczp0ZXN0LGZyZWUsZnVuO2Rlc2NyaXB0aW9uOlRoaXMgaXMgYSB0ZXN0IGRlc2NyaXB0aW9uIGZvciBhbiBhd2Vzb21lIG5mdA==","metadata":"metadata-test","nonEmptyURIs":true}}
`
	require.Equal(t, expectedRes, res[0].String())
}

func TestSerializeAccountsESDTDelete(t *testing.T) {
	t.Parallel()

	accs := map[string]*data.AccountInfo{
		"addr1": {
			Address:    "addr1",
			Nonce:      1,
			TokenName:  "token-0001",
			Properties: "000",
			Balance:    "0",
			BalanceNum: 0,
		},
	}

	res, err := (&accountsProcessor{}).SerializeAccounts(accs, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "delete" : { "_id" : "addr1-token-0001-00" } }
`
	require.Equal(t, expectedRes, res[0].String())
}

func TestSerializeAccountsHistory(t *testing.T) {
	t.Parallel()

	accsh := map[string]*data.AccountBalanceHistory{
		"account1": {
			Address:         "account1",
			Timestamp:       10,
			Balance:         "123",
			Token:           "token-0001",
			IsSender:        true,
			IsSmartContract: true,
		},
	}

	res, err := (&accountsProcessor{}).SerializeAccountsHistory(accsh)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "account1-token-0001-00-10" } }
{"address":"account1","timestamp":10,"balance":"123","token":"token-0001","isSender":true,"isSmartContract":true}
`
	require.Equal(t, expectedRes, res[0].String())
}
