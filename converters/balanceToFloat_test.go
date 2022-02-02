package converters

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeBalanceAsFloat(t *testing.T) {
	t.Parallel()

	ap, _ := NewBalanceConverter(10)
	require.NotNil(t, ap)

	tests := []struct {
		input  *big.Int
		output float64
	}{
		{
			input:  big.NewInt(200000000000000000),
			output: float64(20000000),
		},
		{
			input:  big.NewInt(57777777777),
			output: 5.7777777777,
		},
		{
			input:  big.NewInt(5777779),
			output: 0.0005777779,
		},
		{
			input:  big.NewInt(7),
			output: 0.0000000007,
		},
		{
			input:  big.NewInt(-7),
			output: 0.0,
		},

		{
			input:  big.NewInt(0),
			output: 0.0,
		},
	}

	for _, tt := range tests {
		out := ap.ComputeBalanceAsFloat(tt.input)
		assert.Equal(t, tt.output, out)
	}
}

func TestBigIntToString(t *testing.T) {
	t.Parallel()

	require.Equal(t, "0", BigIntToString(nil))
	require.Equal(t, "0", BigIntToString(big.NewInt(0)))
	require.Equal(t, "1", BigIntToString(big.NewInt(1)))
}
