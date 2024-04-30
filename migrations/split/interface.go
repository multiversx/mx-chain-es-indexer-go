package split

import (
	"bytes"
	"context"
)

type EsClient interface {
	DoBulkRequest(ctx context.Context, buff *bytes.Buffer, index string) error
	DoSearchRequest(ctx context.Context, index string, buff *bytes.Buffer, resBody interface{}) error
	DoMultiGet(ctx context.Context, ids []string, index string, withSource bool, res interface{}) error
}
