package transactions

import (
	"errors"
)

// ErrNilTxHashExtractor signals that a nil tx hash extractor has been provided
var ErrNilTxHashExtractor = errors.New("nil tx hash extractor")

// ErrNilRewardTxDataHandler signals that a nil rewards tx data handler has been provided
var ErrNilRewardTxDataHandler = errors.New("nil reward tx data handler")

// ErrNilIndexTokensHandler signals that a nil index tokens handler has been provided
var ErrNilIndexTokensHandler = errors.New("nil index tokens handler")
