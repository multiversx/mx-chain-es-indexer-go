package data

// Delegator is a structure that is needed to store information about a delegator
type Delegator struct {
	Address        string  `json:"address"`
	Contract       string  `json:"contract"`
	ActiveStake    string  `json:"activeStake"`
	ActiveStakeNum float64 `json:"activeStakeNum"`
	ShouldDelete   bool    `json:"-"`
}
