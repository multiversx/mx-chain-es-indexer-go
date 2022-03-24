package validators

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func TestNewValidatorsProcessor(t *testing.T) {
	t.Parallel()

	vp, err := NewValidatorsProcessor(nil, 0)
	require.Nil(t, vp)
	require.Equal(t, indexer.ErrNilPubkeyConverter, err)
}

func TestValidatorsProcessor_PrepareValidatorsPublicKeys(t *testing.T) {
	t.Parallel()

	vp, _ := NewValidatorsProcessor(&mock.PubkeyConverterMock{}, 0)

	blsKeys := [][]byte{
		[]byte("key1"), []byte("key2"),
	}
	res := vp.PrepareValidatorsPublicKeys(blsKeys)
	require.Equal(t, &data.ValidatorsPublicKeys{
		PublicKeys: []string{
			"6b657931", "6b657932",
		},
	}, res)
}
