package validators

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/stretchr/testify/require"
)

func TestValidatorsProcessor_SerializeValidatorsPubKeys(t *testing.T) {
	t.Parallel()

	validatorsPubKeys := &data.ValidatorsPublicKeys{
		PublicKeys: []string{"bls1", "bls2"},
	}
	buff, err := (&validatorsProcessor{}).SerializeValidatorsPubKeys(validatorsPubKeys)
	require.Nil(t, err)

	expected := `{"publicKeys":["bls1","bls2"]}`
	require.Equal(t, expected, buff.String())
}

func TestValidatorsProcessor_SerializeValidatorsRating(t *testing.T) {
	t.Parallel()

	buff, err := (&validatorsProcessor{}).SerializeValidatorsRating("0", []*data.ValidatorRatingInfo{
		{
			PublicKey: "bls1",
			Rating:    50.1,
		},
		{
			PublicKey: "bls3",
			Rating:    50.2,
		},
	})
	require.Nil(t, err)
	expected := `{ "index" : { "_id" : "bls1_0" } }
{"rating":50.1}
{ "index" : { "_id" : "bls3_0" } }
{"rating":50.2}
`
	require.Equal(t, expected, buff[0].String())
}
