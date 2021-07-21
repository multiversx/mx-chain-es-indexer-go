package transactions

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/stretchr/testify/require"
)

func TestScDeploy_TransactionWithDeploy(t *testing.T) {
	t.Parallel()

	pubKeyConverter, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	scProc := newScDeploysProc(pubKeyConverter, 0)

	res := scProc.searchSCDeployTransactionsOrSCRS([]*data.Transaction{
		{
			Nonce:    1,
			Sender:   "erd1zuv3zd2tuwzj4ktyg5mp4hx0862edcvv8f3xnts5eunnqcmy696q4alcen",
			Receiver: "erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu",
			Hash:     "hash-hash-hash",
			SmartContractResults: []*data.ScResult{
				{
					Sender:   "erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu",
					Receiver: "erd1zuv3zd2tuwzj4ktyg5mp4hx0862edcvv8f3xnts5eunnqcmy696q4alcen",
					Data:     []byte("@6f6b"),
				},
			},
		},
		{},
	}, nil)

	require.Equal(t, &data.ScDeployInfo{
		ScAddress: "erd1qqqqqqqqqqqqqpgqhycd7c2v4pwyqv7d3uppv8m3hw7az327696qyxungp",
		TxHash:    "hash-hash-hash",
		Creator:   "erd1zuv3zd2tuwzj4ktyg5mp4hx0862edcvv8f3xnts5eunnqcmy696q4alcen",
	}, res[0])
}

func TestScDeploy_TransactionCreateDelegationManager(t *testing.T) {
	t.Parallel()

	pubKeyConverter, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	scProc := newScDeploysProc(pubKeyConverter, core.MetachainShardId)

	res := scProc.searchSCDeployTransactionsOrSCRS([]*data.Transaction{
		{
			Nonce:    1,
			Sender:   "erd1wqnzugy8lc67fr4rnalt8guha6sm2tcsnatm2tllq05jh2kr6zlsveq384",
			Receiver: "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqylllslmq6y6",
			Hash:     "hash-hash-hash",
			Data:     []byte("createNewDelegationContract@00@0e"),
			SmartContractResults: []*data.ScResult{
				{
					Sender:   "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqylllslmq6y6",
					Receiver: "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqv8lllls3ydk0k",
					Nonce:    0,
				},
				{
					Nonce:    2,
					Sender:   delegationManagerAddress,
					Receiver: "erd1wqnzugy8lc67fr4rnalt8guha6sm2tcsnatm2tllq05jh2kr6zlsveq384",
					Data:     []byte("@6f6b@0000000000000000000100000000000000000000000000000000000030ffffff"),
				},
				{},
			},
		},
		{},
	}, nil)

	require.Equal(t, &data.ScDeployInfo{
		ScAddress: "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqv8lllls3ydk0k",
		TxHash:    "hash-hash-hash",
		Creator:   "erd1wqnzugy8lc67fr4rnalt8guha6sm2tcsnatm2tllq05jh2kr6zlsveq384",
	}, res[0])
}

func TestScDeploy_SCRDelegationManagerWithDeploy(t *testing.T) {
	t.Parallel()

	pubKeyConverter, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	scProc := newScDeploysProc(pubKeyConverter, core.MetachainShardId)

	res := scProc.searchSCDeployTransactionsOrSCRS(nil, []*data.ScResult{
		{
			Nonce:    1,
			Sender:   "erd1wqnzugy8lc67fr4rnalt8guha6sm2tcsnatm2tllq05jh2kr6zlsveq384",
			Receiver: "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqylllslmq6y6",
			Hash:     "hash-hash-hash",
			Data:     []byte("createNewDelegationContract@00@0e"),
		},
		{
			Sender:   "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqylllslmq6y6",
			Receiver: "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqv8lllls3ydk0k",
			Nonce:    0,
		},
		{
			Nonce:    2,
			Sender:   delegationManagerAddress,
			Receiver: "erd1wqnzugy8lc67fr4rnalt8guha6sm2tcsnatm2tllq05jh2kr6zlsveq384",
			Data:     []byte("@6f6b@0000000000000000000100000000000000000000000000000000000030ffffff"),
		}})

	require.Equal(t, &data.ScDeployInfo{
		ScAddress: "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqv8lllls3ydk0k",
		TxHash:    "hash-hash-hash",
		Creator:   "erd1wqnzugy8lc67fr4rnalt8guha6sm2tcsnatm2tllq05jh2kr6zlsveq384",
	}, res[0])
}

func TestScDeploy_ScrWithDeploy(t *testing.T) {
	t.Parallel()

	pubKeyConverter, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	scProc := newScDeploysProc(pubKeyConverter, 0)

	res := scProc.searchSCDeployTransactionsOrSCRS(nil, []*data.ScResult{
		{},
		{
			Nonce:    1,
			Sender:   "erd1zuv3zd2tuwzj4ktyg5mp4hx0862edcvv8f3xnts5eunnqcmy696q4alcen",
			Receiver: "erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu",
			Hash:     "hash-hash-hash",
		},
		{
			Sender:   "erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu",
			Receiver: "erd1zuv3zd2tuwzj4ktyg5mp4hx0862edcvv8f3xnts5eunnqcmy696q4alcen",
			Data:     []byte("@6f6b"),
		},
	})

	require.Equal(t, &data.ScDeployInfo{
		ScAddress: "erd1qqqqqqqqqqqqqpgqhycd7c2v4pwyqv7d3uppv8m3hw7az327696qyxungp",
		TxHash:    "hash-hash-hash",
		Creator:   "erd1zuv3zd2tuwzj4ktyg5mp4hx0862edcvv8f3xnts5eunnqcmy696q4alcen",
	}, res[0])
}
