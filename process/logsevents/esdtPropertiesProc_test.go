package logsevents

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestRolesProcessorCreateRoleShouldWork(t *testing.T) {
	t.Parallel()

	rolesProc := newRolesProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.BuiltInFunctionSetESDTRole),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), big.NewInt(0).Bytes(), big.NewInt(0).Bytes(), []byte(core.ESDTRoleNFTCreate)},
	}

	tokenRolesAndProperties := data.NewTokenRolesAndProperties()
	rolesProc.processEvent(&argsProcessEvent{
		event:                   event,
		tokenRolesAndProperties: tokenRolesAndProperties,
	})

	expected := map[string][]*data.RoleData{
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

func TestRolesProcessorTransferCreateRole(t *testing.T) {
	t.Parallel()

	rolesProc := newRolesProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.BuiltInFunctionESDTNFTCreateRoleTransfer),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), big.NewInt(0).Bytes(), big.NewInt(0).Bytes(), []byte(strconv.FormatBool(true))},
	}

	tokenRolesAndProperties := data.NewTokenRolesAndProperties()
	rolesProc.processEvent(&argsProcessEvent{
		event:                   event,
		tokenRolesAndProperties: tokenRolesAndProperties,
	})

	expected := map[string][]*data.RoleData{
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
