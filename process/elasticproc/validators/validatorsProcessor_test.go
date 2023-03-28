package validators

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestNewValidatorsProcessor(t *testing.T) {
	t.Parallel()

	vp, err := NewValidatorsProcessor(nil, 0)
	require.Nil(t, vp)
	require.Equal(t, dataindexer.ErrNilPubkeyConverter, err)
}

func TestValidatorsProcessor_PrepareAnSerializeValidatorsPubKeys(t *testing.T) {
	t.Parallel()

	vp, err := NewValidatorsProcessor(&mock.PubkeyConverterMock{}, 0)
	require.Nil(t, err)

	validators := &outport.ValidatorsPubKeys{
		Epoch: 30,
		ShardValidatorsPubKeys: map[uint32]*outport.PubKeys{
			0: {Keys: [][]byte{[]byte("k1"), []byte("k2")}},
		},
	}
	res, err := vp.PrepareAnSerializeValidatorsPubKeys(validators)
	require.Nil(t, err)
	require.Len(t, res, 1)
	require.Equal(t, `{ "index" : { "_id" : "0_30" } }
{"publicKeys":["6b31","6b32"]}
`, res[0].String())
}
