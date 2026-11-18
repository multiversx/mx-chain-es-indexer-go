//go:build integrationtests

package integrationtests

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

// a keyword value bigger than 32766 bytes is rejected by the database and the whole bulk request fails
const maxKeywordLength = 32766

func TestCreateNFTWithOversizedMetadata(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	// every invalid UTF-8 byte is indexed as a character of 3 bytes: without truncation on the indexed value,
	// the name, the tag and the metadata would have 33000 bytes
	invalidUTF8 := strings.Repeat("\xff", 11000)
	// the hash is indexed base64 encoded: without truncation it would have 33336 bytes
	bigHash := bytes.Repeat([]byte("h"), 25000)
	attributes := []byte("tags:" + invalidUTF8 + ";metadata:" + invalidUTF8)

	// the address is in shard 0: the NFT create event is indexed in the tokens index only by the shard of the creator
	address := "erd1ju8pkvg57cwdmjsjx58jlmnuf4l9yspstrhr9tgsrt98n9edpm2qtlgy99"
	esdtToken := &esdt.ESDigitalToken{
		Value:      big.NewInt(1),
		Properties: []byte("3032"),
		TokenMetaData: &esdt.MetaData{
			Name:       []byte(invalidUTF8),
			Creator:    decodeAddress(address),
			Hash:       bigHash,
			Attributes: attributes,
		},
	}
	esdtDataBytes, _ := json.Marshal(esdtToken)

	pool := &outport.TransactionPool{
		Logs: []*transaction.LogData{
			{
				TxHash: hex.EncodeToString([]byte("bigMetadataHash")),
				Log: &transaction.Log{
					Address: decodeAddress(address),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("BIGMETA-abcd"), big.NewInt(1).Bytes(), big.NewInt(1).Bytes(), esdtDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	coreAlteredAccounts := map[string]*alteredAccount.AlteredAccount{
		address: {
			Address: address,
			Balance: "0",
			Tokens: []*alteredAccount.AccountTokenData{
				{
					AdditionalData: &alteredAccount.AdditionalAccountTokenData{
						IsNFTCreate: true,
					},
					Identifier: "BIGMETA-abcd",
					Nonce:      1,
					Balance:    "1",
					Properties: "3032",
					MetaData: &alteredAccount.TokenMetaData{
						Name:       invalidUTF8,
						Creator:    address,
						Hash:       bigHash,
						Attributes: attributes,
					},
				},
			},
		},
	}

	header := &dataBlock.Header{
		Round:     70,
		TimeStamp: 7700,
		ShardID:   0,
	}

	// the block should be indexed: one NFT with huge metadata should not block the indexing
	err = esProc.SaveTransactions(createOutportBlockWithHeader(&dataBlock.Body{}, header, pool, coreAlteredAccounts, testNumOfShards))
	require.NoError(t, err)

	type docWithMetadata struct {
		Data struct {
			Name     string   `json:"name"`
			Hash     string   `json:"hash"`
			Tags     []string `json:"tags"`
			MetaData string   `json:"metadata"`
		} `json:"data"`
	}

	requireMetadataFitsInTheIndex := func(index string, id string) {
		genericResponse := &GenericResponse{}
		errGet := esClient.DoMultiGet(context.Background(), []string{id}, index, true, genericResponse)
		require.Nil(t, errGet)
		require.Len(t, genericResponse.Docs, 1)
		require.True(t, genericResponse.Docs[0].Found, "index %s, id %s", index, id)

		doc := &docWithMetadata{}
		require.Nil(t, json.Unmarshal(genericResponse.Docs[0].Source, doc))

		// 10000 replacement characters of 3 bytes
		expectedText := strings.Repeat("�", 10000)
		require.Equal(t, expectedText, doc.Data.Name, "index %s", index)
		require.Equal(t, expectedText, doc.Data.MetaData, "index %s", index)
		require.Equal(t, []string{expectedText}, doc.Data.Tags, "index %s", index)
		require.LessOrEqual(t, len(doc.Data.Name), maxKeywordLength)

		// the first 22500 bytes of the hash, base64 encoded
		require.Equal(t, 30000, len(doc.Data.Hash), "index %s", index)
		require.LessOrEqual(t, len(doc.Data.Hash), maxKeywordLength)
	}

	requireMetadataFitsInTheIndex(indexerdata.AccountsESDTIndex, fmt.Sprintf("%s-BIGMETA-abcd-01", address))
	requireMetadataFitsInTheIndex(indexerdata.TokensIndex, "BIGMETA-abcd-01")
}
