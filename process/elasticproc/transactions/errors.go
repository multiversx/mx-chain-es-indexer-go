package transactions

import (
	"errors"
)

// ErrNilTxHashExtractor signals that a nil tx hash extractor has been provided
var ErrNilTxHashExtractor = errors.New("nil tx hash extractor")
