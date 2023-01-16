package logsevents

import (
	"encoding/hex"
	"math/big"
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/tokeninfo"
	"github.com/stretchr/testify/require"
)

func TestEsdtPropertiesProcCreateRoleShouldWork(t *testing.T) {
	t.Parallel()

	esdtPropProc := newEsdtPropertiesProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.BuiltInFunctionSetESDTRole),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), big.NewInt(0).Bytes(), big.NewInt(0).Bytes(), []byte(core.ESDTRoleNFTCreate)},
	}

	tokenRolesAndProperties := tokeninfo.NewTokenRolesAndProperties()
	esdtPropProc.processEvent(&argsProcessEvent{
		event:                   event,
		tokenRolesAndProperties: tokenRolesAndProperties,
	})

	expected := map[string][]*tokeninfo.RoleData{
		core.ESDTRoleNFTCreate: {
			{
				Token:   "MYTOKEN-abcd",
				Set:     true,
				Address: "61646472",
			},
		},
	}
	require.Equal(t, expected, tokenRolesAndProperties.GetRoles())
}

func TestEsdtPropertiesProcTransferCreateRole(t *testing.T) {
	t.Parallel()

	esdtPropProc := newEsdtPropertiesProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.BuiltInFunctionESDTNFTCreateRoleTransfer),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), big.NewInt(0).Bytes(), big.NewInt(0).Bytes(), []byte(strconv.FormatBool(true))},
	}

	tokenRolesAndProperties := tokeninfo.NewTokenRolesAndProperties()
	esdtPropProc.processEvent(&argsProcessEvent{
		event:                   event,
		tokenRolesAndProperties: tokenRolesAndProperties,
	})

	expected := map[string][]*tokeninfo.RoleData{
		core.ESDTRoleNFTCreate: {
			{
				Token:   "MYTOKEN-abcd",
				Set:     true,
				Address: "61646472",
			},
		},
	}
	require.Equal(t, expected, tokenRolesAndProperties.GetRoles())
}

func TestEsdtPropertiesProcUpgradeProperties(t *testing.T) {
	t.Parallel()

	esdtPropProc := newEsdtPropertiesProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(upgradePropertiesEvent),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), big.NewInt(0).Bytes(), []byte("canMint"), []byte("true"), []byte("canBurn"), []byte("false")},
	}

	tokenRolesAndProperties := tokeninfo.NewTokenRolesAndProperties()
	esdtPropProc.processEvent(&argsProcessEvent{
		event:                   event,
		tokenRolesAndProperties: tokenRolesAndProperties,
	})

	expected := []*tokeninfo.PropertiesData{
		{
			Token: "MYTOKEN-abcd",
			Properties: map[string]bool{
				"canMint": true,
				"canBurn": false,
			},
		},
	}
	require.Equal(t, expected, tokenRolesAndProperties.GetAllTokensWithProperties())
}

func TestCheckRolesBytes(t *testing.T) {
	t.Parallel()

	role1, _ := hex.DecodeString("01")
	role2, _ := hex.DecodeString("02")
	rolesBytes := [][]byte{role1, role2}
	require.False(t, checkRolesBytes(rolesBytes))

	role1 = []byte("ESDTRoleNFTCreate")
	rolesBytes = [][]byte{role1}
	require.True(t, checkRolesBytes(rolesBytes))
}
