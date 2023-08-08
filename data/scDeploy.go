package data

// ScDeployInfo is the DTO that holds information about a smart contract deployment
type ScDeployInfo struct {
	TxHash        string       `json:"deployTxHash"`
	Creator       string       `json:"deployer"`
	CurrentOwner  string       `json:"currentOwner"`
	CodeHash      []byte       `json:"initialCodeHash"`
	Timestamp     uint64       `json:"timestamp"`
	Upgrades      []*Upgrade   `json:"upgrades"`
	OwnersHistory []*OwnerData `json:"owners"`
}

// Upgrade is the DTO that holds information about a smart contract upgrade
type Upgrade struct {
	TxHash    string `json:"upgradeTxHash"`
	Upgrader  string `json:"upgrader"`
	Timestamp uint64 `json:"timestamp"`
	CodeHash  []byte `json:"codeHash"`
}
