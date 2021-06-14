package accounts

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"time"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
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
const numDecimalsInFloatBalanceESDT = 18

// accountsProcessor a is structure responsible for processing accounts
type accountsProcessor struct {
	dividerForDenomination float64
	balancePrecision       float64
	balancePrecisionESDT   float64
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
		balancePrecisionESDT:   math.Pow(10, float64(numDecimalsInFloatBalanceESDT)),
		dividerForDenomination: math.Pow(10, float64(core.MaxInt(denomination, 0))),
		accountsDB:             accountsDB,
	}, nil
}

// GetAccounts will get accounts for regular operations and esdt operations
func (ap *accountsProcessor) GetAccounts(alteredAccounts data.AlteredAccountsHandler) ([]*data.Account, []*data.AccountESDT) {
	regularAccountsToIndex := make([]*data.Account, 0)
	accountsToIndexESDT := make([]*data.AccountESDT, 0)

	if check.IfNil(alteredAccounts) {
		return regularAccountsToIndex, accountsToIndexESDT
	}

	allAlteredAccounts := alteredAccounts.GetAll()
	for address, altered := range allAlteredAccounts {
		userAccount, err := ap.getUserAccount(address)
		if err != nil || check.IfNil(userAccount) {
			log.Warn("cannot get user account", "address", address, "error", err)
			continue
		}

		regularAccounts, esdtAccounts := splitAlteredAccounts(userAccount, altered)

		regularAccountsToIndex = append(regularAccountsToIndex, regularAccounts...)
		accountsToIndexESDT = append(accountsToIndexESDT, esdtAccounts...)
	}

	return regularAccountsToIndex, accountsToIndexESDT
}

func splitAlteredAccounts(userAccount state.UserAccountHandler, altered []*data.AlteredAccount) ([]*data.Account, []*data.AccountESDT) {
	regularAccountsToIndex := make([]*data.Account, 0)
	accountsToIndexESDT := make([]*data.AccountESDT, 0)
	for _, info := range altered {
		if info.IsESDTOperation || info.IsNFTOperation {
			accountsToIndexESDT = append(accountsToIndexESDT, &data.AccountESDT{
				Account:         userAccount,
				TokenIdentifier: info.TokenIdentifier,
				IsSender:        info.IsSender,
				IsNFTOperation:  info.IsNFTOperation,
				NFTNonce:        info.NFTNonce,
				IsNFTCreate:     info.IsCreate,
				Type:            info.Type,
			})
		}

		ignoreESDTReceiver := (info.IsESDTOperation || info.IsNFTOperation) && !info.IsSender
		if ignoreESDTReceiver {
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
		balanceAsFloat := ap.computeBalanceAsFloat(balance, ap.balancePrecision)
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
func (ap *accountsProcessor) PrepareAccountsMapESDT(
	accounts []*data.AccountESDT,
	timestamp uint64,
) (map[string]*data.AccountInfo, []*data.TokenInfo) {
	accountsESDTMap := make(map[string]*data.AccountInfo)
	tokensCreateInfo := make([]*data.TokenInfo, 0)
	for _, accountESDT := range accounts {
		address := ap.addressPubkeyConverter.Encode(accountESDT.Account.AddressBytes())
		balance, properties, tokenMetaData, err := ap.getESDTInfo(accountESDT)
		if err != nil {
			log.Warn("cannot get esdt info from account",
				"address", address,
				"error", err.Error())
			continue
		}

		acc := &data.AccountInfo{
			Address:         address,
			TokenName:       accountESDT.TokenIdentifier,
			TokenIdentifier: computeTokenIdentifier(accountESDT.TokenIdentifier, accountESDT.NFTNonce),
			TokenNonce:      accountESDT.NFTNonce,
			Balance:         balance.String(),
			BalanceNum:      ap.computeBalanceAsFloat(balance, ap.balancePrecisionESDT),
			Properties:      properties,
			IsSender:        accountESDT.IsSender,
			IsSmartContract: core.IsSmartContractAddress(accountESDT.Account.AddressBytes()),
			MetaData:        tokenMetaData,
		}

		accountsESDTMap[address] = acc

		if !accountESDT.IsNFTOperation || !accountESDT.IsNFTCreate {
			continue
		}

		tokensCreateInfo = append(tokensCreateInfo, &data.TokenInfo{
			Token:      accountESDT.TokenIdentifier,
			Identifier: computeTokenIdentifier(accountESDT.TokenIdentifier, accountESDT.NFTNonce),
			Timestamp:  time.Duration(timestamp),
			MetaData:   tokenMetaData,
			Type:       accountESDT.Type,
		})
	}

	return accountsESDTMap, tokensCreateInfo
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
			Token:           userAccount.TokenName,
			TokenNonce:      userAccount.TokenNonce,
			IsSender:        userAccount.IsSender,
			IsSmartContract: userAccount.IsSmartContract,
			Identifier:      computeTokenIdentifier(userAccount.TokenName, userAccount.TokenNonce),
		}
		addressKey := fmt.Sprintf("%s_%d", address, timestamp)
		accountsMap[addressKey] = acc
	}

	return accountsMap
}

func (ap *accountsProcessor) getESDTInfo(accountESDT *data.AccountESDT) (*big.Int, string, *data.TokenMetaData, error) {
	if accountESDT.TokenIdentifier == "" {
		return big.NewInt(0), "", nil, nil
	}
	if accountESDT.IsNFTOperation && accountESDT.NFTNonce == 0 {
		return big.NewInt(0), "", nil, nil
	}

	tokenKey := []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier + accountESDT.TokenIdentifier)
	if accountESDT.IsNFTOperation {
		nonceBig := big.NewInt(0).SetUint64(accountESDT.NFTNonce)
		tokenKey = append(tokenKey, nonceBig.Bytes()...)
	}

	valueBytes, err := accountESDT.Account.DataTrieTracker().RetrieveValue(tokenKey)
	if err != nil {
		return nil, "", nil, err
	}

	esdtToken := &esdt.ESDigitalToken{}
	err = ap.internalMarshalizer.Unmarshal(esdtToken, valueBytes)
	if err != nil {
		return nil, "", nil, err
	}

	if esdtToken.Value == nil {
		return big.NewInt(0), "", nil, nil
	}

	tokenMetaData := ap.getTokenMetaData(esdtToken)

	return esdtToken.Value, hex.EncodeToString(esdtToken.Properties), tokenMetaData, nil
}

func (ap *accountsProcessor) getTokenMetaData(esdtInfo *esdt.ESDigitalToken) *data.TokenMetaData {
	if esdtInfo.TokenMetaData == nil {
		return nil
	}

	creatorStr := ""
	if esdtInfo.TokenMetaData.Creator != nil {
		creatorStr = ap.addressPubkeyConverter.Encode(esdtInfo.TokenMetaData.Creator)
	}

	return &data.TokenMetaData{
		Name:       string(esdtInfo.TokenMetaData.Name),
		Creator:    creatorStr,
		Royalties:  esdtInfo.TokenMetaData.Royalties,
		Hash:       esdtInfo.TokenMetaData.Hash,
		URIs:       esdtInfo.TokenMetaData.URIs,
		Attributes: data.NewAttributesDTO(esdtInfo.TokenMetaData.Attributes),
	}
}

func (ap *accountsProcessor) computeBalanceAsFloat(balance *big.Int, balancePrecision float64) float64 {
	if balance == nil || balance == big.NewInt(0) {
		return 0
	}

	balanceBigFloat := big.NewFloat(0).SetInt(balance)
	balanceFloat64, _ := balanceBigFloat.Float64()

	bal := balanceFloat64 / ap.dividerForDenomination

	balanceFloatWithDecimals := math.Round(bal*balancePrecision) / balancePrecision

	return core.MaxFloat64(balanceFloatWithDecimals, 0)
}

func computeTokenIdentifier(token string, nonce uint64) string {
	if token == "" || nonce == 0 {
		return ""
	}

	nonceBig := big.NewInt(0).SetUint64(nonce)
	hexEncodedNonce := hex.EncodeToString(nonceBig.Bytes())
	return fmt.Sprintf("%s-%s", token, hexEncodedNonce)
}
