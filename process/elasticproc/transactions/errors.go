package transactions

import (
	"errors"
)

// ErrNilTxHashExtractor signals that a nil tx hash extractor has been provided
var ErrNilTxHashExtractor = errors.New("nil tx hash extractor")

// ErrNilRewardTxData signals that a nil rewards tx data handler has been provided
var ErrNilRewardTxData = errors.New("nil reward tx data handler")
