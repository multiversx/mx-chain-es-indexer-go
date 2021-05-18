package accounts

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-logger/check"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

var log = logger.GetOrCreate("indexer/process/accounts")

const numDecimalsInFloatBalance = 10

// accountsProcessor a is structure responsible for processing accounts
type accountsProcessor struct {
	dividerForDenomination float64
	balancePrecision       float64
	internalMarshalizer    marshal.Marshalizer
	addressPubkeyConverter core.PubkeyConverter
	accountsDB             state.AccountsAdapter
}

// NewAccountsProcessor will create a new instance of accounts processor
func NewAccountsProcessor(
	denomination int,
	marshalizer marshal.Marshalizer,
	addressPubkeyConverter core.PubkeyConverter,
	accountsDB state.AccountsAdapter,
) (*accountsProcessor, error) {
	if denomination < 0 {
		return nil, indexer.ErrNegativeDenominationValue
	}
	if check.IfNil(marshalizer) {
		return nil, indexer.ErrNilMarshalizer
	}
	if check.IfNil(addressPubkeyConverter) {
		return nil, indexer.ErrNilPubkeyConverter
	}
	if check.IfNil(accountsDB) {
		return nil, indexer.ErrNilAccountsDB
	}

	return &accountsProcessor{
		internalMarshalizer:    marshalizer,
		addressPubkeyConverter: addressPubkeyConverter,
		balancePrecision:       math.Pow(10, float64(numDecimalsInFloatBalance)),
		dividerForDenomination: math.Pow(10, float64(core.MaxInt(denomination, 0))),
		accountsDB:             accountsDB,
	}, nil
}

// GetAccounts will get accounts for regular operations and esdt operations
func (ap *accountsProcessor) GetAccounts(alteredAccounts map[string]*data.AlteredAccount) ([]*data.Account, []*data.AccountESDT) {
	regularAccountsToIndex := make([]*data.Account, 0)
	accountsToIndexESDT := make([]*data.AccountESDT, 0)
	for address, info := range alteredAccounts {
		userAccount, err := ap.getUserAccount(address)
		if err != nil {
			log.Warn("cannot get user account", "address", address, "error", err)
			continue
		}

		if info.IsESDTOperation {
			accountsToIndexESDT = append(accountsToIndexESDT, &data.AccountESDT{
				Account:         userAccount,
				TokenIdentifier: info.TokenIdentifier,
				IsSender:        info.IsSender,
			})
		}

		if info.IsESDTOperation && !info.IsSender {
			// should continue because he have an esdt transfer and the current account is not the sender
			// this transfer will not affect the balance of the account
			continue
		}

		regularAccountsToIndex = append(regularAccountsToIndex, &data.Account{
			UserAccount: userAccount,
			IsSender:    info.IsSender,
		})
	}

	return regularAccountsToIndex, accountsToIndexESDT
}

func (ap *accountsProcessor) getUserAccount(address string) (state.UserAccountHandler, error) {
	addressBytes, err := ap.addressPubkeyConverter.Decode(address)
	if err != nil {
		return nil, err
	}

	account, err := ap.accountsDB.LoadAccount(addressBytes)
	if err != nil {
		return nil, err
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		return nil, indexer.ErrCannotCastAccountHandlerToUserAccount
	}

	return userAccount, nil
}

// PrepareRegularAccountsMap will prepare a map of regular accounts
func (ap *accountsProcessor) PrepareRegularAccountsMap(accounts []*data.Account) map[string]*data.AccountInfo {
	accountsMap := make(map[string]*data.AccountInfo)
	for _, userAccount := range accounts {
		balance := userAccount.UserAccount.GetBalance()
		balanceAsFloat := ap.computeBalanceAsFloat(balance)
		acc := &data.AccountInfo{
			Nonce:                    userAccount.UserAccount.GetNonce(),
			Balance:                  balance.String(),
			BalanceNum:               balanceAsFloat,
			IsSender:                 userAccount.IsSender,
			IsSmartContract:          core.IsSmartContractAddress(userAccount.UserAccount.AddressBytes()),
			TotalBalanceWithStake:    balance.String(),
			TotalBalanceWithStakeNum: balanceAsFloat,
		}
		address := ap.addressPubkeyConverter.Encode(userAccount.UserAccount.AddressBytes())
		accountsMap[address] = acc
	}

	return accountsMap
}

// PrepareAccountsMapESDT will prepare a map of accounts with ESDT tokens
func (ap *accountsProcessor) PrepareAccountsMapESDT(accounts []*data.AccountESDT) map[string]*data.AccountInfo {
	accountsESDTMap := make(map[string]*data.AccountInfo)
	for _, accountESDT := range accounts {
		address := ap.addressPubkeyConverter.Encode(accountESDT.Account.AddressBytes())
		balance, properties, err := ap.getESDTInfo(accountESDT)
		if err != nil {
			log.Warn("cannot get esdt info from account",
				"address", address,
				"error", err.Error())
			continue
		}

		acc := &data.AccountInfo{
			Address:         address,
			TokenIdentifier: accountESDT.TokenIdentifier,
			Balance:         balance.String(),
			BalanceNum:      ap.computeBalanceAsFloat(balance),
			Properties:      properties,
			IsSender:        accountESDT.IsSender,
			IsSmartContract: core.IsSmartContractAddress(accountESDT.Account.AddressBytes()),
		}

		accountsESDTMap[address] = acc
	}

	return accountsESDTMap
}

// PrepareAccountsHistory will prepare a map of accounts history balance from a map of accounts
func (ap *accountsProcessor) PrepareAccountsHistory(
	timestamp uint64,
	accounts map[string]*data.AccountInfo,
) map[string]*data.AccountBalanceHistory {
	accountsMap := make(map[string]*data.AccountBalanceHistory)
	for address, userAccount := range accounts {
		acc := &data.AccountBalanceHistory{
			Address:         address,
			Balance:         userAccount.Balance,
			Timestamp:       time.Duration(timestamp),
			TokenIdentifier: userAccount.TokenIdentifier,
			IsSender:        userAccount.IsSender,
			IsSmartContract: userAccount.IsSmartContract,
		}
		addressKey := fmt.Sprintf("%s_%d", address, timestamp)
		accountsMap[addressKey] = acc
	}

	return accountsMap
}

func (ap *accountsProcessor) getESDTInfo(accountESDT *data.AccountESDT) (*big.Int, string, error) {
	if accountESDT.TokenIdentifier == "" {
		return nil, "", nil
	}

	tokenKey := core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier + accountESDT.TokenIdentifier
	valueBytes, err := accountESDT.Account.DataTrieTracker().RetrieveValue([]byte(tokenKey))
	if err != nil {
		return nil, "", err
	}

	esdtToken := &esdt.ESDigitalToken{}
	err = ap.internalMarshalizer.Unmarshal(esdtToken, valueBytes)
	if err != nil {
		return nil, "", err
	}

	return esdtToken.Value, hex.EncodeToString(esdtToken.Properties), nil
}

func (ap *accountsProcessor) computeBalanceAsFloat(balance *big.Int) float64 {
	if balance == nil || balance == big.NewInt(0) {
		return 0
	}

	balanceBigFloat := big.NewFloat(0).SetInt(balance)
	balanceFloat64, _ := balanceBigFloat.Float64()

	bal := balanceFloat64 / ap.dividerForDenomination
	balanceFloatWithDecimals := math.Round(bal*ap.balancePrecision) / ap.balancePrecision

	return core.MaxFloat64(balanceFloatWithDecimals, 0)
}
