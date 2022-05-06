package data

import "gorm.io/gorm"

// ScDeployInfo is the DTO that holds information about a smart contract deployment
type ScDeployInfo struct {
	gorm.Model
	TxHash    string     `json:"deployTxHash"`
	Creator   string     `json:"deployer"`
	Timestamp uint64     `json:"timestamp"`
	Upgrades  []*Upgrade `json:"upgrades" gorm:"foreignKey:ID"`
}

// Upgrade is the DTO that holds information about a smart contract upgrade
type Upgrade struct {
	gorm.Model
	TxHash    string `json:"upgradeTxHash"`
	Upgrader  string `json:"upgrader"`
	Timestamp uint64 `json:"timestamp"`
}
