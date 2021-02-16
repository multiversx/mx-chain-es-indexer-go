package process

import (
	"errors"
)

// ErrNilPubkeyConverter signals that an operation has been attempted to or with a nil public key converter implementation
var ErrNilPubkeyConverter = errors.New("nil pubkey converter")

// ErrNilAccountsDB signals that a nil accounts database has been provided
var ErrNilAccountsDB = errors.New("nil accounts db")

// ErrNilShardCoordinator signals that a nil shard coordinator was provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator")
