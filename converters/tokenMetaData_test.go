package converters

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-vm-common/data/esdt"
	"github.com/stretchr/testify/require"
)

func TestPrepareTokenMetaData(t *testing.T) {
	t.Parallel()

	require.Nil(t, PrepareTokenMetaData(nil, nil))
	require.Nil(t, PrepareTokenMetaData(&mock.PubkeyConverterMock{}, nil))

	require.Equal(t, &data.TokenMetaData{
		Name:       "token",
		Creator:    "63726561746f72",
		Royalties:  0,
		Hash:       []byte("hash"),
		URIs:       nil,
		Attributes: []byte("tags:test,free,fun;description:This is a test description for an awesome nft"),
		Tags:       []string{"test", "free", "fun"},
	}, PrepareTokenMetaData(&mock.PubkeyConverterMock{}, &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			Nonce:      2,
			Name:       []byte("token"),
			Creator:    []byte("creator"),
			Royalties:  0,
			Hash:       []byte("hash"),
			URIs:       nil,
			Attributes: []byte("tags:test,free,fun;description:This is a test description for an awesome nft"),
		},
	}))
}
