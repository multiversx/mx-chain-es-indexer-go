package converters

import (
	"encoding/hex"
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

func TestComputeBalanceToFloat18Decimals(t *testing.T) {
	t.Parallel()

	ap, _ := NewBalanceConverter(18)
	require.NotNil(t, ap)

	require.Equal(t, 1e-18, ap.ComputeESDTBalanceAsFloat(big.NewInt(1)))
	require.Equal(t, 1e-17, ap.ComputeESDTBalanceAsFloat(big.NewInt(10)))
	require.Equal(t, 1e-16, ap.ComputeESDTBalanceAsFloat(big.NewInt(100)))
	require.Equal(t, 1e-15, ap.ComputeESDTBalanceAsFloat(big.NewInt(1000)))
	require.Equal(t, float64(0), ap.ComputeESDTBalanceAsFloat(big.NewInt(0)))
}

func TestComputeBalanceToFloatInf(t *testing.T) {
	t.Parallel()

	ap, _ := NewBalanceConverter(18)
	require.NotNil(t, ap)

	str := "erd1ahmy0yjhjg87n755yv99nzla22zzwfud55sa69gk3anyxyyucq9q2hgxwwerd1ahmy0yjhjg87n755yv99nzla22zzwfud55sa69gk3anyxyyucq9q2hgxwwerd1ahmy0yjhjg87n755yv99nzla22zzwfud55sa69gk3anyxyyucq9q2hgxwwerd1ahmy0yjhjg87n755yv99nzla22zzwfud55sa69gk3anyxyyucq9q2hgxww"
	bigValue := big.NewInt(0).SetBytes([]byte(str))
	require.Equal(t, float64(0), ap.ComputeESDTBalanceAsFloat(bigValue))

	hexValueStr := "2642378914478872274757363306845016200438452904128227930177150600998175785079732885392662259024767727006622197340762976891962082611710440131598510606436851189901116516523843401702254087190199876126823217692111058487892984414016231313689031989"
	decoded, _ := hex.DecodeString(hexValueStr)
	bigValue = big.NewInt(0).SetBytes(decoded)
	require.Equal(t, float64(0), ap.ComputeESDTBalanceAsFloat(bigValue))
}

func TestComputeBalanceToFloatSliceOfValues(t *testing.T) {
	t.Parallel()

	ap, _ := NewBalanceConverter(18)
	require.NotNil(t, ap)

	values := []string{"1000000000000000000", "200000000000000000", "100", "2000", "0"}
	require.Equal(t, []float64{1, 0.2, 1e-16, 2e-15, 0}, ap.ComputeSliceOfStringsAsFloat(values))

	valuesWrong := []string{"wrong"}
	require.Equal(t, []float64{0}, ap.ComputeSliceOfStringsAsFloat(valuesWrong))
}

func TestBigIntToString(t *testing.T) {
	t.Parallel()

	require.Equal(t, "0", BigIntToString(nil))
	require.Equal(t, "0", BigIntToString(big.NewInt(0)))
	require.Equal(t, "1", BigIntToString(big.NewInt(1)))
}
