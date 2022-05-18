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
		Token:      "my-token-1",
		Identifier: "my-token-1-01",
	})
	tokensData.Add(&TokenInfo{
		Token: "my-token-1",
	})

	res := tokensData.GetAllTokens()
	require.Len(t, res, 1)

	res2 := tokensData.GetAll()
	require.Len(t, res2, 2)

	_, found := tokensData.tokensInfo["my-token-1-01"]
	require.True(t, found)
}

func TestTokensInfo_GetAllTokens(t *testing.T) {
	t.Parallel()

	tokensData := NewTokensInfo()

	tokensData.Add(&TokenInfo{
		Token: "my-token-1",
	})
	tokensData.Add(&TokenInfo{
		Token: "my-token-2",
	})
	tokensData.Add(&TokenInfo{
		Token:      "my-token-3",
		Identifier: "my-token-3-03",
	})
	tokensData.Add(&TokenInfo{
		Token:      "my-token-3",
		Identifier: "my-token-3-04",
	})
	tokensData.Add(&TokenInfo{
		Token:      "my-token-3",
		Identifier: "my-token-3-05",
	})

	require.Len(t, tokensData.GetAllTokens(), 3)
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
	tokensData.Add(&TokenInfo{
		Token:      "my-token-3",
		Identifier: "my-token-3-03",
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
			{
				Found: true,
				ID:    "my-token-3",
				Source: SourceToken{
					Type: core.SemiFungibleESDT,
				},
			},
		},
	}

	tokensData.AddTypeAndOwnerFromResponse(res)

	tokenData := tokensData.tokensInfo["my-token-1"]
	require.Equal(t, core.SemiFungibleESDT, tokenData.Type)

	tokenData = tokensData.tokensInfo["my-token-3-03"]
	require.Equal(t, core.SemiFungibleESDT, tokenData.Type)
}
