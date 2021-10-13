package data

type Delegator struct {
	Address        string  `json:"address"`
	Contract       string  `json:"contract"`
	ActiveStake    string  `json:"activeStake"`
	ActiveStakeNum float64 `json:"activeStakeNum"`
}
