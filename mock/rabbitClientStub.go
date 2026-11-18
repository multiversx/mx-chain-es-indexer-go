package mock

import "context"

// RabbitClientStub -
type RabbitClientStub struct {
	PublishCalled func(ctx context.Context, exchange string, messageID string, body []byte) error
	CloseCalled   func() error
}

// Publish -
func (rcs *RabbitClientStub) Publish(ctx context.Context, exchange string, messageID string, body []byte) error {
	if rcs.PublishCalled != nil {
		return rcs.PublishCalled(ctx, exchange, messageID, body)
	}
	return nil
}

// Close -
func (rcs *RabbitClientStub) Close() error {
	if rcs.CloseCalled != nil {
		return rcs.CloseCalled()
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rcs *RabbitClientStub) IsInterfaceNil() bool {
	return rcs == nil
}
