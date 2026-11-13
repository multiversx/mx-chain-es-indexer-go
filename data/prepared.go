package data

import (
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/outport"
)

// PreparedBlock holds the data prepared from a block received from the node.
// The data is prepared only once and then it is handed to every output, the outputs should not change it
type PreparedBlock struct {
	OutportBlock *outport.OutportBlockWithHeader
	// Units holds the data of the header followed by the data of every execution result from the header
	Units []*PreparedBlockUnit
}

// PreparedBlockUnit holds the prepared data of a block: the block from the header or the block of an execution result
type PreparedBlockUnit struct {
	BlockHash []byte
	// Executed is false only for the block of a header v3: the block is not executed yet, its data is received later,
	// as an execution result of another header
	Executed        bool
	HeaderData      *HeaderData
	Results         *PreparedResults
	Logs            *PreparedLogsResults
	OperationsTxs   []*Transaction
	OperationsSCRs  []*ScResult
	AlteredAccounts map[string]*alteredAccount.AlteredAccount
}

// RevertedBlock holds the data of a block that has to be reverted
type RevertedBlock struct {
	Header    coreData.HeaderHandler
	BlockData *outport.BlockData
	// TxHashes and ScrHashes are the hex encoded hashes of the transactions and smart contract results of the block
	TxHashes  []string
	ScrHashes []string
}
