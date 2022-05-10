package data

// ScDeployInfo is the DTO that holds information about a smart contract deployment
type ScDeployInfo struct {
	TxHash    string     `json:"deployTxHash"`
	Creator   string     `json:"deployer"`
	Timestamp uint64     `json:"timestamp"`
	Upgrades  []*Upgrade `json:"upgrades" gorm:"-"`
}

// Upgrade is the DTO that holds information about a smart contract upgrade
type Upgrade struct {
	TxHash    string `json:"upgradeTxHash"`
	Upgrader  string `json:"upgrader"`
	Timestamp uint64 `json:"timestamp"`
}
