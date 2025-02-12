package disabled

import (
	"bytes"
	"context"
)

type elasticClient struct{}

// NewDisabledElasticClient -
func NewDisabledElasticClient() *elasticClient {
	return &elasticClient{}
}

// DoBulkRequest -
func (ec *elasticClient) DoBulkRequest(_ context.Context, _ *bytes.Buffer, _ string) error {
	return nil
}

// DoQueryRemove -
func (ec *elasticClient) DoQueryRemove(_ context.Context, _ string, _ *bytes.Buffer) error {
	return nil
}

// DoMultiGet -
func (ec *elasticClient) DoMultiGet(_ context.Context, _ []string, _ string, _ bool, _ interface{}) error {
	return nil
}

// DoScrollRequest -
func (ec *elasticClient) DoScrollRequest(_ context.Context, _ string, _ []byte, _ bool, _ func(responseBytes []byte) error) error {
	return nil
}

// DoCountRequest -
func (ec *elasticClient) DoCountRequest(_ context.Context, _ string, _ []byte) (uint64, error) {
	return 0, nil
}

// UpdateByQuery -
func (ec *elasticClient) UpdateByQuery(_ context.Context, _ string, _ *bytes.Buffer) error {
	return nil
}

// PutMappings -
func (ec *elasticClient) PutMappings(_ string, _ *bytes.Buffer) error {
	return nil
}

// CheckAndCreateIndex -
func (ec *elasticClient) CheckAndCreateIndex(_ string) error {
	return nil
}

// CheckAndCreateAlias -
func (ec *elasticClient) CheckAndCreateAlias(_ string, _ string) error {
	return nil
}

// CheckAndCreateTemplate -
func (ec *elasticClient) CheckAndCreateTemplate(_ string, _ *bytes.Buffer) error {
	return nil
}

// CheckAndCreatePolicy -
func (ec *elasticClient) CheckAndCreatePolicy(_ string, _ *bytes.Buffer) error {
	return nil
}

// IsEnabled -
func (ec *elasticClient) IsEnabled() bool {
	return false
}

// IsInterfaceNil - returns true if there is no value under the interface
func (ec *elasticClient) IsInterfaceNil() bool {
	return ec == nil
}
