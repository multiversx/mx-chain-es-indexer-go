package data

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/stretchr/testify/require"
)

func TestAlteredAccounts_Add(t *testing.T) {
	t.Parallel()

	altAccounts := NewAlteredAccounts()

	addr := "my-addr"
	acct1 := &AlteredAccount{}
	altAccounts.Add(addr, acct1)

	acct2 := &AlteredAccount{
		IsSender: true,
	}
	altAccounts.Add(addr, acct2)

	res, ok := altAccounts.Get(addr)
	require.True(t, ok)
	require.Equal(t, 1, len(res))
	require.True(t, res[0].IsSender)
}

func TestAlteredAccounts_AddESDT(t *testing.T) {
	t.Parallel()

	altAccounts := NewAlteredAccounts()

	acct1 := &AlteredAccount{}
	addr := "my-addr"
	altAccounts.Add(addr, acct1)

	acct2 := &AlteredAccount{
		TokenIdentifier: "my-token",
		IsESDTOperation: true,
		NFTNonce:        0,
	}
	altAccounts.Add(addr, acct2)

	acct3 := &AlteredAccount{
		IsSender:        true,
		TokenIdentifier: "my-token",
		IsESDTOperation: true,
		NFTNonce:        0,
	}
	altAccounts.Add(addr, acct3)

	acct4 := &AlteredAccount{
		IsSender:        true,
		TokenIdentifier: "my-nft-token",
		IsNFTOperation:  true,
		NFTNonce:        1,
		Type:            core.NonFungibleESDT,
	}
	altAccounts.Add(addr, acct4)

	acct5 := &AlteredAccount{
		IsSender:        true,
		TokenIdentifier: "my-nft-token",
		IsNFTOperation:  true,
		NFTNonce:        1,
		Type:            core.NonFungibleESDT,
	}
	altAccounts.Add(addr, acct5)

	acct6 := &AlteredAccount{
		IsSender:        true,
		TokenIdentifier: "my-nft-token",
		IsNFTOperation:  true,
		NFTNonce:        2,
		Type:            core.NonFungibleESDT,
	}
	altAccounts.Add(addr, acct6)

	require.Equal(t, 1, altAccounts.Len())
	res, ok := altAccounts.Get(addr)
	require.True(t, ok)
	require.Equal(t, 3, len(res))
	require.Equal(t, &AlteredAccount{
		IsSender:        true,
		IsESDTOperation: true,
		TokenIdentifier: "my-token",
	}, res[0])

	require.Equal(t, &AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: "my-nft-token",
		NFTNonce:        1,
		Type:            core.NonFungibleESDT,
	}, res[1])

	require.Equal(t, &AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: "my-nft-token",
		NFTNonce:        2,
		Type:            core.NonFungibleESDT,
	}, res[2])
}

func TestAlteredAccounts_GetAll(t *testing.T) {
	t.Parallel()

	altAccounts := &alteredAccounts{}

	res := altAccounts.GetAll()
	require.NotNil(t, res)

	altAccounts = NewAlteredAccounts()
	res = altAccounts.GetAll()
	require.NotNil(t, res)
}
