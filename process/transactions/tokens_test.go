package transactions

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/vm"
	"github.com/stretchr/testify/require"
)

func TestNewTokensProcessor_SearchIssueFungibleESDTTransactions(t *testing.T) {
	t.Parallel()

	pubKeyConverter := &mock.PubkeyConverterMock{}

	esdtSCAddr := pubKeyConverter.Encode(core.ESDTSCAddress)

	sender := "erd1yp046t9pc009mkaxyws5dm434ydut348va5ypjd3ng57euy3yjkqcnrfh3"
	txs := []*data.Transaction{
		{
			Sender:   sender,
			Receiver: esdtSCAddr,
			Nonce:    15,
			Data:     []byte("issue@446f6c6c6172@555344@0ba43b7400@05@63616e55706772616465@74727565"),
			SmartContractResults: []*data.ScResult{
				{
					Nonce:    16,
					Sender:   esdtSCAddr,
					Receiver: sender,
					Data:     []byte("@6f6b"),
				},
				{
					Nonce:    0,
					Sender:   esdtSCAddr,
					Receiver: sender,
					Data:     []byte("ESDTTransfer@5553442d326665643930@0ba43b7400"),
				},
			},
		},
	}

	tokensProc := newTokensProcessor(core.MetachainShardId, pubKeyConverter)
	tokensInfo := tokensProc.searchForTokenIssueTransactions(txs, 3000)
	require.Equal(t, &data.TokenInfo{
		Name:      "Dollar",
		Ticker:    "USD",
		Token:     "USD-2fed90",
		Issuer:    "erd1yp046t9pc009mkaxyws5dm434ydut348va5ypjd3ng57euy3yjkqcnrfh3",
		Type:      core.FungibleESDT,
		Timestamp: time.Duration(3000),
	}, tokensInfo[0])
}

func TestNewTokensProcessor_SearchIssueSemiFungibleESDTTransactions(t *testing.T) {
	t.Parallel()

	pubKeyConverter := &mock.PubkeyConverterMock{}

	esdtSCAddr := pubKeyConverter.Encode(core.ESDTSCAddress)

	sender := "erd1yp046t9pc009mkaxyws5dm434ydut348va5ypjd3ng57euy3yjkqcnrfh3"
	txs := []*data.Transaction{
		{
			Sender:   sender,
			Receiver: esdtSCAddr,
			Nonce:    6,
			Data:     []byte("issueSemiFungible@53656d6946756e6769626c65546f6b656e@534654@63616e467265657a65@74727565@63616e57697065@74727565@63616e5061757365@74727565@63616E5472616E736665724E4654437265617465526F6C65@74727565"),
			SmartContractResults: []*data.ScResult{
				{
					Nonce:    7,
					Sender:   esdtSCAddr,
					Receiver: sender,
					Data:     []byte("@6f6b@5346542d623264303139"),
				},
			},
		},
	}

	tokensProc := newTokensProcessor(core.MetachainShardId, pubKeyConverter)
	tokensInfo := tokensProc.searchForTokenIssueTransactions(txs, 3000)
	require.Equal(t, &data.TokenInfo{
		Name:      "SemiFungibleToken",
		Ticker:    "SFT",
		Token:     "SFT-b2d019",
		Issuer:    "erd1yp046t9pc009mkaxyws5dm434ydut348va5ypjd3ng57euy3yjkqcnrfh3",
		Type:      core.SemiFungibleESDT,
		Timestamp: time.Duration(3000),
	}, tokensInfo[0])
}

func TestNewTokensProcessor_SearchIssueNonFungibleESDTTransactions(t *testing.T) {
	t.Parallel()

	pubKeyConverter := &mock.PubkeyConverterMock{}

	esdtSCAddr := pubKeyConverter.Encode(core.ESDTSCAddress)

	sender := "erd1yp046t9pc009mkaxyws5dm434ydut348va5ypjd3ng57euy3yjkqcnrfh3"
	txs := []*data.Transaction{
		{
			Sender:   sender,
			Receiver: esdtSCAddr,
			Nonce:    25,
			Data:     []byte("issueNonFungible@4d794e4654@4e4654@63616e467265657a65@74727565@63616e57697065@74727565@63616e5061757365@74727565@63616E5472616E736665724E4654437265617465526F6C65@74727565"),
			SmartContractResults: []*data.ScResult{
				{
					Nonce:    26,
					Sender:   esdtSCAddr,
					Receiver: sender,
					Data:     []byte("@6f6b@4e46542d393437346262"),
				},
			},
		},
	}

	tokensProc := newTokensProcessor(core.MetachainShardId, pubKeyConverter)
	tokensInfo := tokensProc.searchForTokenIssueTransactions(txs, 3000)
	require.Equal(t, &data.TokenInfo{
		Name:      "MyNFT",
		Ticker:    "NFT",
		Token:     "NFT-9474bb",
		Issuer:    "erd1yp046t9pc009mkaxyws5dm434ydut348va5ypjd3ng57euy3yjkqcnrfh3",
		Type:      core.NonFungibleESDT,
		Timestamp: time.Duration(3000),
	}, tokensInfo[0])
}

func TestNewTokensProcessor_SearchIssueFungibleESDTSCResults(t *testing.T) {
	t.Parallel()

	pubKeyConverter := &mock.PubkeyConverterMock{}

	esdtSCAddr := pubKeyConverter.Encode(core.ESDTSCAddress)

	sender := "erd1yp046t9pc009mkaxyws5dm434ydut348va5ypjd3ng57euy3yjkqcnrfh3"
	scrs := []*data.ScResult{
		{
			Sender:   sender,
			Receiver: esdtSCAddr,
			Nonce:    15,
			Data:     []byte("issue@446f6c6c6172@555344@0ba43b7400@05@63616e55706772616465@74727565"),
		},
		{
			Nonce:    0,
			Sender:   esdtSCAddr,
			Receiver: sender,
			Data:     []byte("ESDTTransfer@5553442d326665643930@0ba43b7400"),
			CallType: fmt.Sprintf("%d", vm.AsynchronousCallBack),
		},
	}

	tokensProc := newTokensProcessor(core.MetachainShardId, pubKeyConverter)
	tokensInfo := tokensProc.searchForTokenIssueScrs(scrs, 3000)
	require.Equal(t, &data.TokenInfo{
		Name:      "Dollar",
		Ticker:    "USD",
		Token:     "USD-2fed90",
		Issuer:    "erd1yp046t9pc009mkaxyws5dm434ydut348va5ypjd3ng57euy3yjkqcnrfh3",
		Type:      core.FungibleESDT,
		Timestamp: time.Duration(3000),
	}, tokensInfo[0])
}

func TestNewTokensProcessor_SearchIssueSemiFungibleESDTScResults(t *testing.T) {
	t.Parallel()

	pubKeyConverter := &mock.PubkeyConverterMock{}

	esdtSCAddr := pubKeyConverter.Encode(core.ESDTSCAddress)

	sender := "erd1yp046t9pc009mkaxyws5dm434ydut348va5ypjd3ng57euy3yjkqcnrfh3"
	scrs := []*data.ScResult{
		{
			Sender:   sender,
			Receiver: esdtSCAddr,
			Nonce:    6,
			Data:     []byte("issueSemiFungible@53656d6946756e6769626c65546f6b656e@534654@63616e467265657a65@74727565@63616e57697065@74727565@63616e5061757365@74727565@63616E5472616E736665724E4654437265617465526F6C65@74727565"),
		},
		{
			Nonce:    7,
			Sender:   esdtSCAddr,
			Receiver: sender,
			Data:     []byte("@00@5346542d623264303139"),
		},
	}

	tokensProc := newTokensProcessor(core.MetachainShardId, pubKeyConverter)
	tokensInfo := tokensProc.searchForTokenIssueScrs(scrs, 3000)
	require.Equal(t, &data.TokenInfo{
		Name:      "SemiFungibleToken",
		Ticker:    "SFT",
		Token:     "SFT-b2d019",
		Issuer:    "erd1yp046t9pc009mkaxyws5dm434ydut348va5ypjd3ng57euy3yjkqcnrfh3",
		Type:      core.SemiFungibleESDT,
		Timestamp: time.Duration(3000),
	}, tokensInfo[0])
}
