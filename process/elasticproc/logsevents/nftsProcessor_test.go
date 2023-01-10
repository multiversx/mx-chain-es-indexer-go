package logsevents

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
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

	nftsProc := newNFTsProcessor(&mock.PubkeyConverterMock{}, &mock.MarshalizerMock{})

	tokensCreateInfo := data.NewTokensInfo()
	res := nftsProc.processEvent(&argsProcessEvent{
		event:       event,
		tokens:      tokensCreateInfo,
		timestamp:   1000,
		selfShardID: 2,
		numOfShards: 3,
	})
	require.Equal(t, true, res.processed)
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

func TestNftsProcessor_processLogAndEventsNFTs_Wipe(t *testing.T) {
	t.Parallel()

	nonce := uint64(20)
	nftsProc := newNFTsProcessor(&mock.PubkeyConverterMock{}, &mock.MarshalizerMock{})

	events := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.BuiltInFunctionESDTWipe),
		Topics:     [][]byte{[]byte("nft-0123"), big.NewInt(0).SetUint64(nonce).Bytes(), big.NewInt(1).Bytes(), []byte("receiver")},
	}

	tokensSupply := data.NewTokensInfo()
	res := nftsProc.processEvent(&argsProcessEvent{
		event:        events,
		timestamp:    10000,
		tokensSupply: tokensSupply,
		numOfShards:  3,
		selfShardID:  2,
	})
	require.Equal(t, true, res.processed)
	require.Equal(t, &data.TokenInfo{
		Identifier: "nft-0123-14",
		Token:      "nft-0123",
		Nonce:      20,
		Timestamp:  time.Duration(10000),
	}, tokensSupply.GetAll()[0])
}
