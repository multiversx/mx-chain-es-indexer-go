package validators

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/stretchr/testify/require"
)

func TestValidatorsProcessor_SerializeValidatorsRating(t *testing.T) {
	t.Parallel()

	ratingInfo := &outport.ValidatorsRating{
		ShardID: 0,
		Epoch:   0,
		ValidatorsRatingInfo: []*outport.ValidatorRatingInfo{
			{
				PublicKey: "bls1",
				Rating:    50.1,
			},
			{
				PublicKey: "bls3",
				Rating:    50.2,
			},
		},
	}
	buff, err := (&validatorsProcessor{}).SerializeValidatorsRating(ratingInfo)
	require.Nil(t, err)
	expected := `{ "index" : { "_id" : "bls1_0" } }
{"rating":50.1}
{ "index" : { "_id" : "bls3_0" } }
{"rating":50.2}
`
	require.Equal(t, expected, buff[0].String())
}
