package accounts

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/stretchr/testify/require"
)

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

func TestSerializeAccountsESDT(t *testing.T) {
	t.Parallel()

	accs := map[string]*data.AccountInfo{
		"addr1": {
			Address:         "addr1",
			Nonce:           1,
			TokenIdentifier: "token-0001",
			Properties:      "000",
			TokenNonce:      5,
			Balance:         "10000000000000",
			BalanceNum:      1,
		},
	}

	res, err := (&accountsProcessor{}).SerializeAccounts(accs, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "addr1_token-0001_5" } }
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
			TokenIdentifier: "token-0001",
			Properties:      "000",
			TokenNonce:      5,
			Balance:         "10000000000000",
			BalanceNum:      1,
			MetaData: &data.TokenMetaData{
				Name:      "nft",
				Creator:   "010101",
				Royalties: 1,
				Hash:      []byte("hash"),
				URIs: [][]byte{
					[]byte("uri"),
				},
				Attributes: []byte("atr"),
			},
		},
	}

	res, err := (&accountsProcessor{}).SerializeAccounts(accs, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "addr1_token-0001_5" } }
{"address":"addr1","nonce":1,"balance":"10000000000000","balanceNum":1,"token":"token-0001","tokenNonce":5,"properties":"000","tokenMetaData":{"name":"nft","creator":"010101","royalties":1,"hash":"aGFzaA==","uris":["dXJp"],"attributes":"YXRy"}}
`
	require.Equal(t, expectedRes, res[0].String())
}

func TestSerializeAccountsESDTDelete(t *testing.T) {
	t.Parallel()

	accs := map[string]*data.AccountInfo{
		"addr1": {
			Address:         "addr1",
			Nonce:           1,
			TokenIdentifier: "token-0001",
			Properties:      "000",
			Balance:         "0",
			BalanceNum:      0,
		},
	}

	res, err := (&accountsProcessor{}).SerializeAccounts(accs, true)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "delete" : { "_id" : "addr1_token-0001_0" } }
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
			TokenIdentifier: "token-0001",
			IsSender:        true,
			IsSmartContract: true,
		},
	}

	res, err := (&accountsProcessor{}).SerializeAccountsHistory(accsh)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { } }
{"address":"account1","timestamp":10,"balance":"123","token":"token-0001","isSender":true,"isSmartContract":true}
`
	require.Equal(t, expectedRes, res[0].String())
}
