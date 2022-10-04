package logsevents

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestNftsProcessor_processLogAndEventsNFTs(t *testing.T) {
	t.Parallel()

	esdtData := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			Creator: []byte("creator"),
		},
	}
	esdtDataBytes, _ := json.Marshal(esdtData)

	nonce := uint64(19)
	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
		Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(nonce).Bytes(), big.NewInt(1).Bytes(), esdtDataBytes},
	}

	nftsProc := newNFTsProcessor(&mock.ShardCoordinatorMock{}, &mock.PubkeyConverterMock{}, &mock.MarshalizerMock{})

	altered := data.NewAlteredAccounts()

	tokensCreateInfo := data.NewTokensInfo()
	res := nftsProc.processEvent(&argsProcessEvent{
		event:     event,
		accounts:  altered,
		tokens:    tokensCreateInfo,
		timestamp: 1000,
	})
	require.Equal(t, "my-token-13", res.identifier)
	require.Equal(t, "1", res.value)
	require.Equal(t, true, res.processed)

	alteredAddr, ok := altered.Get("61646472")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: "my-token",
		NFTNonce:        19,
		IsNFTCreate:     true,
	}, alteredAddr[0])

	require.Equal(t, &data.TokenInfo{
		Identifier: "my-token-13",
		Token:      "my-token",
		Timestamp:  1000,
		Issuer:     "",
		Nonce:      uint64(19),
		Data: &data.TokenMetaData{
			Creator: hex.EncodeToString([]byte("creator")),
		},
	}, tokensCreateInfo.GetAll()[0])

}

func TestNftsProcessor_processLogAndEventsNFTs_TransferNFT(t *testing.T) {
	t.Parallel()

	nonce := uint64(19)
	nftsProc := newNFTsProcessor(&mock.ShardCoordinatorMock{}, &mock.PubkeyConverterMock{}, &mock.MarshalizerMock{})

	events := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.BuiltInFunctionESDTNFTTransfer),
		Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(nonce).Bytes(), big.NewInt(1).Bytes(), []byte("receiver")},
	}

	altered := data.NewAlteredAccounts()

	res := nftsProc.processEvent(&argsProcessEvent{
		event:     events,
		accounts:  altered,
		timestamp: 10000,
	})
	require.Equal(t, "my-token-13", res.identifier)
	require.Equal(t, "1", res.value)
	require.Equal(t, true, res.processed)

	alteredAddrSender, ok := altered.Get("61646472")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: "my-token",
		NFTNonce:        19,
	}, alteredAddrSender[0])

	alteredAddrReceiver, ok := altered.Get("7265636569766572")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: "my-token",
		NFTNonce:        19,
	}, alteredAddrReceiver[0])
}

func TestNftsProcessor_processLogAndEventsNFTs_Wipe(t *testing.T) {
	t.Parallel()

	nonce := uint64(20)
	nftsProc := newNFTsProcessor(&mock.ShardCoordinatorMock{}, &mock.PubkeyConverterMock{}, &mock.MarshalizerMock{})

	events := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.BuiltInFunctionESDTWipe),
		Topics:     [][]byte{[]byte("nft-0123"), big.NewInt(0).SetUint64(nonce).Bytes(), big.NewInt(1).Bytes(), []byte("receiver")},
	}

	altered := data.NewAlteredAccounts()

	res := nftsProc.processEvent(&argsProcessEvent{
		event:        events,
		accounts:     altered,
		timestamp:    10000,
		tokensSupply: data.NewTokensInfo(),
	})
	require.Equal(t, "nft-0123-14", res.identifier)
	require.Equal(t, "1", res.value)
	require.Equal(t, true, res.processed)

	alteredAddrSender, ok := altered.Get("61646472")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: "nft-0123",
		NFTNonce:        20,
	}, alteredAddrSender[0])

	alteredAddrReceiver, ok := altered.Get("7265636569766572")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: "nft-0123",
		NFTNonce:        20,
	}, alteredAddrReceiver[0])
}
