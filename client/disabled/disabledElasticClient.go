package disabled

import (
	"bytes"
	"context"
)

type disabledElasticClient struct{}

// NewDisabledElasticClient -
func NewDisabledElasticClient() *disabledElasticClient {
	return &disabledElasticClient{}
}

// DoBulkRequest -
func (dec *disabledElasticClient) DoBulkRequest(_ context.Context, _ *bytes.Buffer, _ string) error {
	return nil
}

// DoQueryRemove -
func (dec *disabledElasticClient) DoQueryRemove(_ context.Context, _ string, _ *bytes.Buffer) error {
	return nil
}

// DoMultiGet -
func (dec *disabledElasticClient) DoMultiGet(_ context.Context, _ []string, _ string, _ bool, _ interface{}) error {
	return nil
}

// DoScrollRequest -
func (dec *disabledElasticClient) DoScrollRequest(_ context.Context, _ string, _ []byte, _ bool, _ func(responseBytes []byte) error) error {
	return nil
}

// DoCountRequest -
func (dec *disabledElasticClient) DoCountRequest(_ context.Context, _ string, _ []byte) (uint64, error) {
	return 0, nil
}

// UpdateByQuery -
func (dec *disabledElasticClient) UpdateByQuery(_ context.Context, _ string, _ *bytes.Buffer) error {
	return nil
}

// PutMappings -
func (dec *disabledElasticClient) PutMappings(_ string, _ *bytes.Buffer) error {
	return nil
}

// CheckAndCreateIndex -
func (dec *disabledElasticClient) CheckAndCreateIndex(_ string) error {
	return nil
}

// CheckAndCreateAlias -
func (dec *disabledElasticClient) CheckAndCreateAlias(_ string, _ string) error {
	return nil
}

// CheckAndCreateTemplate -
func (dec *disabledElasticClient) CheckAndCreateTemplate(_ string, _ *bytes.Buffer) error {
	return nil
}

// CheckAndCreatePolicy -
func (dec *disabledElasticClient) CheckAndCreatePolicy(_ string, _ *bytes.Buffer) error {
	return nil
}

// IsEnabled -
func (dec *disabledElasticClient) IsEnabled() bool {
	return false
}

// IsInterfaceNil - returns true if there is no value under the interface
func (dec *disabledElasticClient) IsInterfaceNil() bool {
	return dec == nil
}
