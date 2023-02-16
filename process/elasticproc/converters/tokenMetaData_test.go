package converters

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/stretchr/testify/require"
)

func TestPrepareTokenMetaData(t *testing.T) {
	t.Parallel()

	require.Nil(t, PrepareTokenMetaData(nil))

	expectedTokenMetaData := &data.TokenMetaData{
		Name:               "token",
		Creator:            "creator",
		Royalties:          0,
		Hash:               []byte("hash"),
		URIs:               [][]byte{[]byte("https://ipfs.io/ipfs/something"), []byte("uri")},
		Attributes:         []byte("tags:test,free,fun;description:This is a test description for an awesome nft;metadata:metadata-test"),
		Tags:               []string{"test", "free", "fun"},
		MetaData:           "metadata-test",
		NonEmptyURIs:       true,
		WhiteListedStorage: true,
	}

	result := PrepareTokenMetaData(&outport.TokenMetaData{
		Nonce:      2,
		Name:       "token",
		Creator:    "creator",
		Royalties:  0,
		Hash:       []byte("hash"),
		URIs:       [][]byte{[]byte(ipfsURL + "something"), []byte("uri")},
		Attributes: []byte("tags:test,free,fun;description:This is a test description for an awesome nft;metadata:metadata-test"),
	})

	require.Equal(t, expectedTokenMetaData, result)
}

func TestPrepareNFTUpdateData(t *testing.T) {
	t.Parallel()

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)

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
	err := PrepareNFTUpdateData(buffSlice, nftUpdateData, false, "tokens")
	require.Nil(t, err)
	require.Equal(t, `{"update":{ "_index":"tokens","_id":"MYTKN-abcd-01"}}
{"script": {"source": "if (ctx._source.containsKey('data')) {ctx._source.data.attributes = params.attributes;if (!params.metadata.isEmpty() ) {ctx._source.data.metadata = params.metadata} else {if (ctx._source.data.containsKey('metadata')) {ctx._source.data.remove('metadata')}}if (params.tags != null) {ctx._source.data.tags = params.tags} else {if (ctx._source.data.containsKey('tags')) {ctx._source.data.remove('tags')}}}","lang": "painless","params": {"attributes": "YWFhYQ==", "metadata": "", "tags": null}}, "upsert": {}}
{"update":{ "_index":"tokens","_id":"TOKEN-1234-1a"}}
{"script": {"source": "if (ctx._source.containsKey('data')) {if (!ctx._source.data.containsKey('uris')) {ctx._source.data.uris = params.uris;} else {int i;for ( i = 0; i < params.uris.length; i++) {boolean found = false;int j;for ( j = 0; j < ctx._source.data.uris.length; j++) {if ( params.uris.get(i) == ctx._source.data.uris.get(j) ) {found = true;break}}if ( !found ) {ctx._source.data.uris.add(params.uris.get(i))}}}ctx._source.data.nonEmptyURIs = true;}","lang": "painless","params": {"uris": ["dXJpMQ==","dXJpMg=="]}},"upsert": {}}
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
