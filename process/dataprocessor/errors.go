package dataprocessor

import "errors"

// ErrNilTransactionsHandler signals that a nil transactions handler has been provided
var ErrNilTransactionsHandler = errors.New("nil transactions handler")

// ErrNilLogsAndEventsHandler signals that a nil logs and events handler has been provided
var ErrNilLogsAndEventsHandler = errors.New("nil logs and events handler")

// ErrNilOperationsHandler signals that a nil operations handler has been provided
var ErrNilOperationsHandler = errors.New("nil operations handler")

// ErrNoOutputs signals that no output has been provided
var ErrNoOutputs = errors.New("no outputs provided")

// ErrNilOutput signals that a nil output has been provided
var ErrNilOutput = errors.New("nil output")
