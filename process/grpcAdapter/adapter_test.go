package grpcAdapter

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/outport/grpcadapter"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

var errExpected = errors.New("expected error")

func TestNewGRPCAdapter_NilHandler(t *testing.T) {
	t.Parallel()

	adapter, err := NewGRPCAdapter(nil)
	require.Equal(t, grpcadapter.ErrNilOutportServiceHandler, err)
	require.Nil(t, adapter)
}

func TestNewGRPCAdapter_ShouldWork(t *testing.T) {
	t.Parallel()

	adapter, err := NewGRPCAdapter(&mock.DataIndexerStub{})
	require.Nil(t, err)
	require.NotNil(t, adapter)
}

func TestGrpcAdapter_SaveBlock(t *testing.T) {
	t.Parallel()

	called := false
	adapter, err := NewGRPCAdapter(&mock.DataIndexerStub{
		SaveBlockCalled: func(outportBlock *outport.OutportBlock) error {
			called = true
			return nil
		},
	})
	require.Nil(t, err)

	response, err := adapter.SaveBlock(context.Background(), &outport.OutportBlock{})
	require.Nil(t, err)
	require.NotNil(t, response)
	require.True(t, called)
}

func TestGrpcAdapter_SaveBlockErr(t *testing.T) {
	t.Parallel()

	adapter, err := NewGRPCAdapter(&mock.DataIndexerStub{
		SaveBlockCalled: func(outportBlock *outport.OutportBlock) error {
			return errExpected
		},
	})
	require.Nil(t, err)

	response, err := adapter.SaveBlock(context.Background(), &outport.OutportBlock{})
	require.Equal(t, errExpected, err)
	require.NotNil(t, response)
}

func TestGrpcAdapter_RevertIndexedBlock(t *testing.T) {
	t.Parallel()

	called := false
	adapter, err := NewGRPCAdapter(&mock.DataIndexerStub{
		RevertIndexedBlockCalled: func(blockData *outport.BlockData) error {
			called = true
			return nil
		},
	})
	require.Nil(t, err)

	response, err := adapter.RevertIndexedBlock(context.Background(), &outport.BlockData{})
	require.Nil(t, err)
	require.NotNil(t, response)
	require.True(t, called)
}

func TestGrpcAdapter_RevertIndexedBlockErr(t *testing.T) {
	t.Parallel()

	adapter, err := NewGRPCAdapter(&mock.DataIndexerStub{
		RevertIndexedBlockCalled: func(blockData *outport.BlockData) error {
			return errExpected
		},
	})
	require.Nil(t, err)

	response, err := adapter.RevertIndexedBlock(context.Background(), &outport.BlockData{})
	require.Equal(t, errExpected, err)
	require.NotNil(t, response)
}

func TestGrpcAdapter_SaveRoundsInfo(t *testing.T) {
	t.Parallel()

	called := false
	adapter, err := NewGRPCAdapter(&mock.DataIndexerStub{
		SaveRoundsInfoCalled: func(roundsInfos *outport.RoundsInfo) error {
			called = true
			return nil
		},
	})
	require.Nil(t, err)

	response, err := adapter.SaveRoundsInfo(context.Background(), &outport.RoundsInfo{})
	require.Nil(t, err)
	require.NotNil(t, response)
	require.True(t, called)
}

func TestGrpcAdapter_SaveRoundsInfoErr(t *testing.T) {
	t.Parallel()

	adapter, err := NewGRPCAdapter(&mock.DataIndexerStub{
		SaveRoundsInfoCalled: func(roundsInfos *outport.RoundsInfo) error {
			return errExpected
		},
	})
	require.Nil(t, err)

	response, err := adapter.SaveRoundsInfo(context.Background(), &outport.RoundsInfo{})
	require.Equal(t, errExpected, err)
	require.NotNil(t, response)
}

func TestGrpcAdapter_SaveValidatorsPubKeys(t *testing.T) {
	t.Parallel()

	called := false
	adapter, err := NewGRPCAdapter(&mock.DataIndexerStub{
		SaveValidatorsPubKeysCalled: func(validatorsPubKeys *outport.ValidatorsPubKeys) error {
			called = true
			return nil
		},
	})
	require.Nil(t, err)

	response, err := adapter.SaveValidatorsPubKeys(context.Background(), &outport.ValidatorsPubKeys{})
	require.Nil(t, err)
	require.NotNil(t, response)
	require.True(t, called)
}

func TestGrpcAdapter_SaveValidatorsPubKeysErr(t *testing.T) {
	t.Parallel()

	adapter, err := NewGRPCAdapter(&mock.DataIndexerStub{
		SaveValidatorsPubKeysCalled: func(validatorsPubKeys *outport.ValidatorsPubKeys) error {
			return errExpected
		},
	})
	require.Nil(t, err)

	response, err := adapter.SaveValidatorsPubKeys(context.Background(), &outport.ValidatorsPubKeys{})
	require.Equal(t, errExpected, err)
	require.NotNil(t, response)
}

func TestGrpcAdapter_SaveValidatorsRating(t *testing.T) {
	t.Parallel()

	called := false
	adapter, err := NewGRPCAdapter(&mock.DataIndexerStub{
		SaveValidatorsRatingCalled: func(ratingData *outport.ValidatorsRating) error {
			called = true
			return nil
		},
	})
	require.Nil(t, err)

	response, err := adapter.SaveValidatorsRating(context.Background(), &outport.ValidatorsRating{})
	require.Nil(t, err)
	require.NotNil(t, response)
	require.True(t, called)
}

func TestGrpcAdapter_SaveValidatorsRatingErr(t *testing.T) {
	t.Parallel()

	adapter, err := NewGRPCAdapter(&mock.DataIndexerStub{
		SaveValidatorsRatingCalled: func(ratingData *outport.ValidatorsRating) error {
			return errExpected
		},
	})
	require.Nil(t, err)

	response, err := adapter.SaveValidatorsRating(context.Background(), &outport.ValidatorsRating{})
	require.Equal(t, errExpected, err)
	require.NotNil(t, response)
}

func TestGrpcAdapter_SaveAccounts(t *testing.T) {
	t.Parallel()

	called := false
	adapter, err := NewGRPCAdapter(&mock.DataIndexerStub{
		SaveAccountsCalled: func(accountsData *outport.Accounts) error {
			called = true
			return nil
		},
	})
	require.Nil(t, err)

	response, err := adapter.SaveAccounts(context.Background(), &outport.Accounts{})
	require.Nil(t, err)
	require.NotNil(t, response)
	require.True(t, called)
}

func TestGrpcAdapter_SaveAccountsErr(t *testing.T) {
	t.Parallel()

	adapter, err := NewGRPCAdapter(&mock.DataIndexerStub{
		SaveAccountsCalled: func(accountsData *outport.Accounts) error {
			return errExpected
		},
	})
	require.Nil(t, err)

	response, err := adapter.SaveAccounts(context.Background(), &outport.Accounts{})
	require.Equal(t, errExpected, err)
	require.NotNil(t, response)
}

func TestGrpcAdapter_FinalizedBlockEvent(t *testing.T) {
	t.Parallel()

	called := false
	adapter, err := NewGRPCAdapter(&mock.DataIndexerStub{
		FinalizedBlockCalled: func(finalizedBlock *outport.FinalizedBlock) error {
			called = true
			return nil
		},
	})
	require.Nil(t, err)

	response, err := adapter.FinalizedBlockEvent(context.Background(), &outport.FinalizedBlock{})
	require.Nil(t, err)
	require.NotNil(t, response)
	require.True(t, called)
}

func TestGrpcAdapter_FinalizedBlockEventErr(t *testing.T) {
	t.Parallel()

	adapter, err := NewGRPCAdapter(&mock.DataIndexerStub{
		FinalizedBlockCalled: func(finalizedBlock *outport.FinalizedBlock) error {
			return errExpected
		},
	})
	require.Nil(t, err)

	response, err := adapter.FinalizedBlockEvent(context.Background(), &outport.FinalizedBlock{})
	require.Equal(t, errExpected, err)
	require.NotNil(t, response)
}

func TestGrpcAdapter_createResponse(t *testing.T) {
	t.Parallel()

	adapter, err := NewGRPCAdapter(&mock.DataIndexerStub{})
	require.Nil(t, err)

	ga, ok := adapter.(*grpcAdapter)
	require.True(t, ok)

	t.Run("should return response when call succeeds", func(t *testing.T) {
		response, err := ga.createResponse(func() error {
			return nil
		})
		require.Nil(t, err)
		require.NotNil(t, response)
	})

	t.Run("should return response and propagate error when call fails", func(t *testing.T) {
		response, err := ga.createResponse(func() error {
			return fmt.Errorf("some error")
		})
		require.EqualError(t, err, "some error")
		require.NotNil(t, response)
	})
}

func TestGrpcAdapter_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var ga *grpcAdapter
	require.True(t, ga.IsInterfaceNil())

	ga = &grpcAdapter{}
	require.False(t, ga.IsInterfaceNil())
}
