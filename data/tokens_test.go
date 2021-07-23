package data

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/stretchr/testify/require"
)

func TestTokensInfo_AddGet(t *testing.T) {
	t.Parallel()

	tokensData := NewTokensInfo()

	tokensData.Add(&TokenInfo{
		Token: "my-token-1",
	})
	tokensData.Add(&TokenInfo{
		Token: "my-token-2",
	})

	res := tokensData.GetAllTokens()
	require.Len(t, res, 2)

	res2 := tokensData.GetAll()
	require.Len(t, res2, 2)
}

func TestTokensInfo_AddTypeFromResponse(t *testing.T) {
	t.Parallel()

	tokensData := NewTokensInfo()

	tokensData.Add(&TokenInfo{
		Token: "my-token-1",
	})
	tokensData.Add(&TokenInfo{
		Token: "my-token-2",
	})

	res := &ResponseTokens{
		Docs: []ResponseTokenDB{
			{
				Found: true,
				ID:    "my-token-1",
				Source: SourceToken{
					Type: core.SemiFungibleESDT,
				},
			},
			{
				Found: false,
			},
		},
	}

	tokensData.AddTypeFromResponse(res)

	tokenData := tokensData.tokensInfo["my-token-1"]
	require.Equal(t, core.SemiFungibleESDT, tokenData.Type)
}
