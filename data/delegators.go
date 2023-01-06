package data

import "time"

// Delegator is a structure that is needed to store information about a delegator
type Delegator struct {
	Address         string        `json:"address"`
	Contract        string        `json:"contract"`
	Timestamp       time.Duration `json:"timestamp"`
	ActiveStake     string        `json:"activeStake"`
	ActiveStakeNum  float64       `json:"activeStakeNum"`
	ShouldDelete    bool          `json:"-"`
	UnDelegateInfo  *UnDelegate   `json:"-"`
	WithdrawFundIDs []string      `json:"-"`
}

// UnDelegate is a structure that is needed to store information about user unDelegate position
type UnDelegate struct {
	Timestamp time.Duration `json:"timestamp"`
	ID        string        `json:"id"`
	Value     string        `json:"value"`
	ValueNum  float64       `json:"valueNum"`
}
