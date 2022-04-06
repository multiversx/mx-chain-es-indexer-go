package accounts

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math/big"
	"testing"
	"time"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	coreIndexerData "github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/stretchr/testify/require"
)

var balanceConverter, _ = converters.NewBalanceConverter(10)

func TestNewAccountsProcessor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		argsFunc func() (marshal.Marshalizer, core.PubkeyConverter, indexer.BalanceConverter)
		exError  error
	}{
		{
			name: "NilBalanceConverter",
			argsFunc: func() (marshal.Marshalizer, core.PubkeyConverter, indexer.BalanceConverter) {
				return &mock.MarshalizerMock{}, &mock.PubkeyConverterMock{}, nil
			},
			exError: indexer.ErrNilBalanceConverter,
		},
		{
			name: "NilMarshalizer",
			argsFunc: func() (marshal.Marshalizer, core.PubkeyConverter, indexer.BalanceConverter) {
				return nil, &mock.PubkeyConverterMock{}, balanceConverter
			},
			exError: indexer.ErrNilMarshalizer,
		},
		{
			name: "NilPubKeyConverter",
			argsFunc: func() (marshal.Marshalizer, core.PubkeyConverter, indexer.BalanceConverter) {
				return &mock.MarshalizerMock{}, nil, balanceConverter
			},
			exError: indexer.ErrNilPubkeyConverter,
		},
		{
			name: "ShouldWork",
			argsFunc: func() (marshal.Marshalizer, core.PubkeyConverter, indexer.BalanceConverter) {
				return &mock.MarshalizerMock{}, &mock.PubkeyConverterMock{}, balanceConverter
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAccountsProcessor(tt.argsFunc())
			require.True(t, errors.Is(err, tt.exError))
		})
	}
}

func TestAccountsProcessor_GetAccountsWithNil(t *testing.T) {
	t.Parallel()

	ap, _ := NewAccountsProcessor(&mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), balanceConverter)

	regularAccounts, esdtAccounts := ap.GetAccounts(nil, nil)
	require.Len(t, regularAccounts, 0)
	require.Len(t, esdtAccounts, 0)
}

func TestAccountsProcessor_PrepareRegularAccountsMapWithNil(t *testing.T) {
	t.Parallel()

	ap, _ := NewAccountsProcessor(&mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), balanceConverter)

	accountsInfo := ap.PrepareRegularAccountsMap(0, nil)
	require.Len(t, accountsInfo, 0)
}

func TestGetESDTInfo(t *testing.T) {
	t.Parallel()

	ap, _ := NewAccountsProcessor(&mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), balanceConverter)
	require.NotNil(t, ap)

	tokenIdentifier := "token-001"
	wrapAccount := &data.AccountESDT{
		Account: &coreIndexerData.AlteredAccount{
			Address: "",
			Balance: "1000",
			Nonce:   0,
			Tokens: []*coreIndexerData.AccountTokenData{
				{
					Identifier: tokenIdentifier,
					Balance:    "1000",
					Properties: "ok",
				},
			},
		},
		TokenIdentifier: tokenIdentifier,
	}
	balance, prop, _, err := ap.getESDTInfo(wrapAccount)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(1000), balance)
	require.Equal(t, hex.EncodeToString([]byte("ok")), prop)
}

func TestGetESDTInfoNFT(t *testing.T) {
	t.Parallel()

	ap, _ := NewAccountsProcessor(&mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), balanceConverter)
	require.NotNil(t, ap)

	tokenIdentifier := "token-001"
	wrapAccount := &data.AccountESDT{
		Account: &coreIndexerData.AlteredAccount{
			Address: "",
			Balance: "1",
			Nonce:   10,
			Tokens: []*coreIndexerData.AccountTokenData{
				{
					Identifier: tokenIdentifier,
					Balance:    "1",
					Nonce:      10,
					Properties: "ok",
				},
			},
		},
		TokenIdentifier: tokenIdentifier,
		IsNFTOperation:  true,
		NFTNonce:        10,
	}
	balance, prop, _, err := ap.getESDTInfo(wrapAccount)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(1), balance)
	require.Equal(t, hex.EncodeToString([]byte("ok")), prop)
}

func TestGetESDTInfoNFTWithMetaData(t *testing.T) {
	t.Parallel()

	pubKeyConverter := mock.NewPubkeyConverterMock(32)
	ap, _ := NewAccountsProcessor(&mock.MarshalizerMock{}, pubKeyConverter, balanceConverter)
	require.NotNil(t, ap)

	nftName := "Test-nft"
	creator := []byte("010101")

	tokenIdentifier := "token-001"
	wrapAccount := &data.AccountESDT{
		Account: &coreIndexerData.AlteredAccount{
			Address: "",
			Balance: "1",
			Nonce:   1,
			Tokens: []*coreIndexerData.AccountTokenData{
				{
					Identifier: tokenIdentifier,
					Balance:    "1",
					Properties: "ok",
					Nonce:      10,
					MetaData: &esdt.MetaData{
						Nonce:     10,
						Name:      []byte(nftName),
						Creator:   creator,
						Royalties: 2,
					},
				},
			},
		},
		TokenIdentifier: tokenIdentifier,
		IsNFTOperation:  true,
		NFTNonce:        10,
	}
	balance, prop, metaData, err := ap.getESDTInfo(wrapAccount)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(1), balance)
	require.Equal(t, hex.EncodeToString([]byte("ok")), prop)
	require.Equal(t, &data.TokenMetaData{
		Name:      nftName,
		Creator:   pubKeyConverter.Encode(creator),
		Royalties: 2,
	}, metaData)
}

func TestAccountsProcessor_GetAccountsEGLDAccounts(t *testing.T) {
	t.Parallel()

	addr := "aaaabbbb"
	acc := &coreIndexerData.AlteredAccount{}
	alteredAccountsMap := map[string]*coreIndexerData.AlteredAccount{
		addr: acc,
	}
	ap, _ := NewAccountsProcessor(&mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), balanceConverter)
	require.NotNil(t, ap)

	alteredAccounts := data.NewAlteredAccounts()
	alteredAccounts.Add(addr, &data.AlteredAccount{
		IsESDTOperation: false,
		TokenIdentifier: "",
	})

	accounts, esdtAccounts := ap.GetAccounts(alteredAccounts, alteredAccountsMap)
	require.Equal(t, 0, len(esdtAccounts))
	require.Equal(t, []*data.Account{
		{
			UserAccount: acc,
		},
	}, accounts)
}

func TestAccountsProcessor_GetAccountsESDTAccount(t *testing.T) {
	t.Parallel()

	addr := "aaaabbbb"
	acc := &coreIndexerData.AlteredAccount{
		Address: addr,
		Balance: "1",
	}
	alteredAccountsMap := map[string]*coreIndexerData.AlteredAccount{
		addr: acc,
	}
	ap, _ := NewAccountsProcessor(&mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), balanceConverter)
	require.NotNil(t, ap)

	alteredAccounts := data.NewAlteredAccounts()
	alteredAccounts.Add(addr, &data.AlteredAccount{
		IsESDTOperation: true,
		TokenIdentifier: "token",
	})
	accounts, esdtAccounts := ap.GetAccounts(alteredAccounts, alteredAccountsMap)
	require.Equal(t, 0, len(accounts))
	require.Equal(t, []*data.AccountESDT{
		{Account: acc, TokenIdentifier: "token"},
	}, esdtAccounts)
}

func TestAccountsProcessor_GetAccountsESDTAccountNewAccountShouldBeInRegularAccounts(t *testing.T) {
	t.Parallel()

	addr := "aaaabbbb"
	acc := &coreIndexerData.AlteredAccount{}
	alteredAccountsMap := map[string]*coreIndexerData.AlteredAccount{
		addr: acc,
	}
	ap, _ := NewAccountsProcessor(&mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), balanceConverter)
	require.NotNil(t, ap)

	alteredAccounts := data.NewAlteredAccounts()
	alteredAccounts.Add(addr, &data.AlteredAccount{
		IsESDTOperation: true,
		TokenIdentifier: "token",
	})
	accounts, esdtAccounts := ap.GetAccounts(alteredAccounts, alteredAccountsMap)
	require.Equal(t, 1, len(accounts))
	require.Equal(t, []*data.AccountESDT{
		{Account: acc, TokenIdentifier: "token"},
	}, esdtAccounts)

	require.Equal(t, []*data.Account{
		{UserAccount: acc, IsSender: false},
	}, accounts)
}

func TestAccountsProcessor_PrepareAccountsMapEGLD(t *testing.T) {
	t.Parallel()

	addrBytes := bytes.Repeat([]byte{0}, 32)
	addr := hex.EncodeToString(addrBytes)

	account := &coreIndexerData.AlteredAccount{
		Address: addr,
		Balance: "1000",
		Nonce:   1,
	}

	egldAccount := &data.Account{
		UserAccount: account,
		IsSender:    false,
	}

	ap, _ := NewAccountsProcessor(&mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), balanceConverter)
	require.NotNil(t, ap)

	res := ap.PrepareRegularAccountsMap(123, []*data.Account{egldAccount})
	require.Equal(t, &data.AccountInfo{
		Address:                  addr,
		Nonce:                    1,
		Balance:                  "1000",
		BalanceNum:               balanceConverter.ComputeBalanceAsFloat(big.NewInt(1000)),
		TotalBalanceWithStake:    "1000",
		TotalBalanceWithStakeNum: balanceConverter.ComputeBalanceAsFloat(big.NewInt(1000)),
		IsSmartContract:          true,
			Timestamp:                time.Duration(123),
	},
		res[addr])
}

func TestAccountsProcessor_PrepareAccountsMapESDT(t *testing.T) {
	t.Parallel()

	addr := "aaaabbbb"

	account := &coreIndexerData.AlteredAccount{
		Address: hex.EncodeToString([]byte(addr)),
		Tokens: []*coreIndexerData.AccountTokenData{
			{
				Balance:    "1000",
				Identifier: "token",
				Nonce:      15,
				Properties: "ok",
				MetaData: &esdt.MetaData{
					Creator: []byte("creator"),
				},
			},
			{
				Balance:    "1000",
				Identifier: "token",
				Nonce:      16,
				Properties: "ok",
				MetaData: &esdt.MetaData{
					Creator: []byte("creator"),
				},
			},
		},
	}
	ap, _ := NewAccountsProcessor(&mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), balanceConverter)
	require.NotNil(t, ap)

	accountsESDT := []*data.AccountESDT{
		{Account: account, TokenIdentifier: "token", IsNFTOperation: true, NFTNonce: 15},
		{Account: account, TokenIdentifier: "token", IsNFTOperation: true, NFTNonce: 16},
	}
	res, _ := ap.PrepareAccountsMapESDT(123, accountsESDT)
	require.Len(t, res, 2)

	require.Equal(t, &data.AccountInfo{
		Address:         hex.EncodeToString([]byte(addr)),
		Balance:         "1000",
		BalanceNum:      balanceConverter.ComputeBalanceAsFloat(big.NewInt(1000)),
		TokenName:       "token",
		TokenIdentifier: "token-0f",
		Properties:      hex.EncodeToString([]byte("ok")),
		TokenNonce:      15,
		Data: &data.TokenMetaData{
			Creator: "63726561746f72",
		},
		Timestamp: time.Duration(123),
	}, res[hex.EncodeToString([]byte(addr))+"-token-15"])

	require.Equal(t, &data.AccountInfo{
		Address:         hex.EncodeToString([]byte(addr)),
		Balance:         "1000",
		BalanceNum:      balanceConverter.ComputeBalanceAsFloat(big.NewInt(1000)),
		TokenName:       "token",
		TokenIdentifier: "token-10",
		Properties:      hex.EncodeToString([]byte("ok")),
		TokenNonce:      16,
		Data: &data.TokenMetaData{
			Creator: "63726561746f72",
		},
		Timestamp: time.Duration(123),
	}, res[hex.EncodeToString([]byte(addr))+"-token-16"])
}

func TestAccountsProcessor_PrepareAccountsHistory(t *testing.T) {
	t.Parallel()

	accounts := map[string]*data.AccountInfo{
		"addr1": {
			Address:    "addr1",
			Balance:    "112",
			TokenName:  "token-112",
			TokenNonce: 10,
			IsSender:   true,
		},
	}

	ap, _ := NewAccountsProcessor(&mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), balanceConverter)

	res := ap.PrepareAccountsHistory(100, accounts)
	accountBalanceHistory := res["addr1-token-112-10"]
	require.Equal(t, &data.AccountBalanceHistory{
		Address:    "addr1",
		Timestamp:  100,
		Balance:    "112",
		Token:      "token-112",
		IsSender:   true,
		TokenNonce: 10,
		Identifier: "token-112-0a",
	}, accountBalanceHistory)
}
