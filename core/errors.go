package core

import "errors"

// ErrNilMetricsHandler signals that a nil metrics handler has been provided
var ErrNilMetricsHandler = errors.New("nil metrics handler")

// ErrNilFacadeHandler signal that a nil facade handler has been provided
var ErrNilFacadeHandler = errors.New("nil facade handler")
