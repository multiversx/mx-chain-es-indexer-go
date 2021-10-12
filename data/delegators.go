package data

type Delegator struct {
	Address        string  `json:"-"`
	Contract       string  `json:"contract"`
	ActiveStake    string  `json:"activeStake"`
	ActiveStakeNum float64 `json:"activeStakeNum"`
}
