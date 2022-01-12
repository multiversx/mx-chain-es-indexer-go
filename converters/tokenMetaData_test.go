package converters

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/stretchr/testify/require"
)

func TestPrepareTokenMetaData(t *testing.T) {
	t.Parallel()

	require.Nil(t, PrepareTokenMetaData(nil, nil))
	require.Nil(t, PrepareTokenMetaData(&mock.PubkeyConverterMock{}, nil))

	expectedTokenMetaData := &data.TokenMetaData{
		Name:               "token",
		Creator:            "63726561746f72",
		Royalties:          0,
		Hash:               []byte("hash"),
		URIs:               [][]byte{[]byte("https://ipfs.io/ipfs/something"), []byte("uri")},
		Attributes:         []byte("tags:test,free,fun;description:This is a test description for an awesome nft;metadata:metadata-test"),
		Tags:               []string{"test", "free", "fun"},
		MetaData:           "metadata-test",
		NonEmptyURIs:       true,
		WhiteListedStorage: true,
	}

	result := PrepareTokenMetaData(&mock.PubkeyConverterMock{}, &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			Nonce:      2,
			Name:       []byte("token"),
			Creator:    []byte("creator"),
			Royalties:  0,
			Hash:       []byte("hash"),
			URIs:       [][]byte{[]byte(ipfsURL + "something"), []byte("uri")},
			Attributes: []byte("tags:test,free,fun;description:This is a test description for an awesome nft;metadata:metadata-test"),
		},
	})

	require.Equal(t, expectedTokenMetaData, result)
}

func TestPrepareNFTUpdateData(t *testing.T) {
	t.Parallel()

	buffSlice := data.NewBufferSlice()

	nftUpdateData := []*data.NFTDataUpdate{
		{
			Identifier:    "MYTKN-abcd-01",
			NewAttributes: []byte("aaaa"),
		},
		{
			Identifier: "TOKEN-1234-1a",
			URIsToAdd:  [][]byte{[]byte("uri1"), []byte("uri2")},
		},
	}
	err := PrepareNFTUpdateData(buffSlice, nftUpdateData, false)
	require.Nil(t, err)
	require.Equal(t, `{"update":{"_id":"MYTKN-abcd-01", "_type": "_doc"}}
{"script": {"source": "if (ctx._source.containsKey('data')) {ctx._source.data.attributes = params.attributes}","lang": "painless","params": {"attributes": "YWFhYQ=="}}, "upsert": {}}
{"update":{"_id":"TOKEN-1234-1a", "_type": "_doc"}}
{"script": {"source": "if (ctx._source.containsKey('data')) { if (!ctx._source.data.containsKey('uris')) { ctx._source.data.uris = params.uris; } else {  ctx._source.data.uris.addAll(params.uris); }}","lang": "painless","params": {"uris": ["dXJpMQ==","dXJpMg=="]}},"upsert": {}}
`, buffSlice.Buffers()[0].String())
}

func TestWhiteListedStorage(t *testing.T) {
	t.Parallel()

	uris := [][]byte{[]byte("https://my-test-nft.pinata.cloud/ipfs/aaaaaa")}
	require.True(t, whiteListedStorage(uris))

	uris = [][]byte{[]byte("ipfs://my-test-nft")}
	require.True(t, whiteListedStorage(uris))

	uris = [][]byte{[]byte("https://dweb.link/ipfs/my-test-nft")}
	require.True(t, whiteListedStorage(uris))

	uris = [][]byte{[]byte("http://dweb.link/ipfs/my-test-nft")}
	require.False(t, whiteListedStorage(uris))

	uris = [][]byte{[]byte("https://dwb.link/ipfs/my-test-nft")}
	require.False(t, whiteListedStorage(uris))
}
