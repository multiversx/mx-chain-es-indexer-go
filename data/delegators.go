package data

// Delegator is a structure that is needed to store information about a delegator
type Delegator struct {
	Address         string      `json:"address"`
	Contract        string      `json:"contract"`
	Timestamp       uint64      `json:"timestamp"`
	TimestampMs     uint64      `json:"timestampMs,omitempty"`
	ActiveStake     string      `json:"activeStake"`
	ActiveStakeNum  float64     `json:"activeStakeNum"`
	ShouldDelete    bool        `json:"-"`
	UnDelegateInfo  *UnDelegate `json:"-"`
	WithdrawFundIDs []string    `json:"-"`
}

// UnDelegate is a structure that is needed to store information about user unDelegate position
type UnDelegate struct {
	Timestamp   uint64  `json:"timestamp"`
	TimestampMs uint64  `json:"timestampMs,omitempty"`
	ID          string  `json:"id"`
	Value       string  `json:"value"`
	ValueNum    float64 `json:"valueNum"`
}
