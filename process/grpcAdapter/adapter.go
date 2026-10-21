package grpcAdapter

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/outport/grpcadapter"
)

type grpcAdapter struct {
	handler DataIndexer
}

func NewGRPCAdapter(handler DataIndexer) (outport.OutportServiceServer, error) {
	if check.IfNil(handler) {
		return nil, grpcadapter.ErrNilOutportServiceHandler
	}

	return &grpcAdapter{
		handler: handler,
	}, nil
}

func (ga *grpcAdapter) SaveBlock(_ context.Context, in *outport.OutportBlock) (*outport.ResponseData, error) {
	return ga.createResponse(func() error {
		return ga.handler.SaveBlock(in)
	})
}

func (ga *grpcAdapter) RevertIndexedBlock(_ context.Context, in *outport.BlockData) (*outport.ResponseData, error) {
	return ga.createResponse(func() error {
		return ga.handler.RevertIndexedBlock(in)
	})
}

func (ga *grpcAdapter) SaveRoundsInfo(_ context.Context, in *outport.RoundsInfo) (*outport.ResponseData, error) {
	return ga.createResponse(func() error {
		return ga.handler.SaveRoundsInfo(in)
	})
}

func (ga *grpcAdapter) SaveValidatorsPubKeys(_ context.Context, in *outport.ValidatorsPubKeys) (*outport.ResponseData, error) {
	return ga.createResponse(func() error {
		return ga.handler.SaveValidatorsPubKeys(in)
	})
}

func (ga *grpcAdapter) SaveValidatorsRating(_ context.Context, in *outport.ValidatorsRating) (*outport.ResponseData, error) {
	return ga.createResponse(func() error {
		return ga.handler.SaveValidatorsRating(in)
	})
}

func (ga *grpcAdapter) SaveAccounts(_ context.Context, in *outport.Accounts) (*outport.ResponseData, error) {
	return ga.createResponse(func() error {
		return ga.handler.SaveAccounts(in)
	})
}

func (ga *grpcAdapter) FinalizedBlockEvent(_ context.Context, in *outport.FinalizedBlock) (*outport.ResponseData, error) {
	return ga.createResponse(func() error {
		return ga.handler.FinalizedBlock(in)
	})
}

func (ga *grpcAdapter) IsInterfaceNil() bool {
	return ga == nil
}

func (ga *grpcAdapter) createResponse(call func() error) (*outport.ResponseData, error) {
	start := time.Now()
	err := call()

	return &outport.ResponseData{
		IndexingTimeInMs: uint64(time.Since(start).Milliseconds()),
	}, err
}
