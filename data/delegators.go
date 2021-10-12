package data

type Delegator struct {
	Contract       string  `json:"contract"`
	ActiveStake    string  `json:"activeStake"`
	ActiveStakeNum float64 `json:"activeStakeNum"`
}
