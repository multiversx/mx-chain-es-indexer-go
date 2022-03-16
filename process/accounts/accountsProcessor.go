package accounts

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	coreIndexerData "github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("indexer/process/accounts")

// accountsProcessor a is structure responsible for processing accounts
type accountsProcessor struct {
	internalMarshalizer    marshal.Marshalizer
	addressPubkeyConverter core.PubkeyConverter
	balanceConverter       indexer.BalanceConverter
}

// NewAccountsProcessor will create a new instance of accounts processor
func NewAccountsProcessor(
	marshalizer marshal.Marshalizer,
	addressPubkeyConverter core.PubkeyConverter,
	balanceConverter indexer.BalanceConverter,
) (*accountsProcessor, error) {
	if check.IfNil(marshalizer) {
		return nil, indexer.ErrNilMarshalizer
	}
	if check.IfNil(addressPubkeyConverter) {
		return nil, indexer.ErrNilPubkeyConverter
	}
	if check.IfNil(balanceConverter) {
		return nil, indexer.ErrNilBalanceConverter
	}

	return &accountsProcessor{
		internalMarshalizer:    marshalizer,
		addressPubkeyConverter: addressPubkeyConverter,
		balanceConverter:       balanceConverter,
	}, nil
}

// GetAccounts will get accounts for regular operations and esdt operations
func (ap *accountsProcessor) GetAccounts(alteredAccounts data.AlteredAccountsHandler, coreAlteredAccounts map[string]*coreIndexerData.AlteredAccount) ([]*data.Account, []*data.AccountESDT) {
	regularAccountsToIndex := make([]*data.Account, 0)
	accountsToIndexESDT := make([]*data.AccountESDT, 0)

	if check.IfNil(alteredAccounts) {
		return regularAccountsToIndex, accountsToIndexESDT
	}

	allAlteredAccounts := alteredAccounts.GetAll()
	for address, altered := range allAlteredAccounts {
		alteredAccount := coreAlteredAccounts[address]
		if alteredAccount == nil {
			log.Warn("account not found in core altered accounts map", "address", address)
			continue
		}

		regularAccounts, esdtAccounts := splitAlteredAccounts(alteredAccount, altered)

		regularAccountsToIndex = append(regularAccountsToIndex, regularAccounts...)
		accountsToIndexESDT = append(accountsToIndexESDT, esdtAccounts...)
	}

	return regularAccountsToIndex, accountsToIndexESDT
}

func splitAlteredAccounts(
	account *coreIndexerData.AlteredAccount,
	altered []*data.AlteredAccount,
) ([]*data.Account, []*data.AccountESDT) {
	regularAccountsToIndex := make([]*data.Account, 0)
	accountsToIndexESDT := make([]*data.AccountESDT, 0)
	for _, info := range altered {
		if info.IsESDTOperation || info.IsNFTOperation {
			accountsToIndexESDT = append(accountsToIndexESDT, &data.AccountESDT{
				Account:         account,
				TokenIdentifier: info.TokenIdentifier,
				IsSender:        info.IsSender,
				IsNFTOperation:  info.IsNFTOperation,
				NFTNonce:        info.NFTNonce,
				Type:            info.Type,
			})
		}

		// if the balance of the ESDT receiver is 0 the receiver is a new account most probably, and we should index it
		ignoreReceiver := !info.BalanceChange && notZeroBalance(account.Balance) && !info.IsSender
		if ignoreReceiver {
			continue
		}

		regularAccountsToIndex = append(regularAccountsToIndex, &data.Account{
			UserAccount: account,
			IsSender:    info.IsSender,
		})
	}

	return regularAccountsToIndex, accountsToIndexESDT
}

func notZeroBalance(balance string) bool {
	return len(balance) > 0 && balance != "0"
}

// PrepareRegularAccountsMap will prepare a map of regular accounts
func (ap *accountsProcessor) PrepareRegularAccountsMap(accounts []*data.Account) map[string]*data.AccountInfo {
	accountsMap := make(map[string]*data.AccountInfo)
	for _, userAccount := range accounts {
		address := userAccount.UserAccount.Address
		addressBytes, err := ap.addressPubkeyConverter.Decode(address)
		if err != nil {
			log.Warn("PrepareRegularAccountsMap: cannot decode address", "address", address, "error", err)
			continue
		}
		balance, ok := big.NewInt(0).SetString(userAccount.UserAccount.Balance, 10)
		if !ok {
			log.Warn("cannot cast account's balance to big int", "value", userAccount.UserAccount.Balance)
			continue
		}

		balanceAsFloat := ap.balanceConverter.ComputeBalanceAsFloat(balance)
		acc := &data.AccountInfo{
			Address:                  address,
			Nonce:                    userAccount.UserAccount.Nonce,
			Balance:                  balance.String(),
			BalanceNum:               balanceAsFloat,
			IsSender:                 userAccount.IsSender,
			IsSmartContract:          core.IsSmartContractAddress(addressBytes),
			TotalBalanceWithStake:    balance.String(),
			TotalBalanceWithStakeNum: balanceAsFloat,
		}

		accountsMap[address] = acc
	}

	return accountsMap
}

// PrepareAccountsMapESDT will prepare a map of accounts with ESDT tokens
func (ap *accountsProcessor) PrepareAccountsMapESDT(
	accounts []*data.AccountESDT,
) map[string]*data.AccountInfo {
	accountsESDTMap := make(map[string]*data.AccountInfo)
	for _, accountESDT := range accounts {
		address := accountESDT.Account.Address
		addressBytes, err := ap.addressPubkeyConverter.Decode(address)
		if err != nil {
			log.Warn("PrepareAccountsMapESDT: cannot decode address", "address", address, "error", err)
			continue
		}
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
			TokenIdentifier: converters.ComputeTokenIdentifier(accountESDT.TokenIdentifier, accountESDT.NFTNonce),
			TokenNonce:      accountESDT.NFTNonce,
			Balance:         balance.String(),
			BalanceNum:      ap.balanceConverter.ComputeESDTBalanceAsFloat(balance),
			Properties:      properties,
			IsSender:        accountESDT.IsSender,
			IsSmartContract: core.IsSmartContractAddress(addressBytes),
			Data:            tokenMetaData,
		}

		keyInMap := fmt.Sprintf("%s-%s-%d", acc.Address, acc.TokenName, accountESDT.NFTNonce)
		accountsESDTMap[keyInMap] = acc
	}

	return accountsESDTMap
}

// PrepareAccountsHistory will prepare a map of accounts history balance from a map of accounts
func (ap *accountsProcessor) PrepareAccountsHistory(
	timestamp uint64,
	accounts map[string]*data.AccountInfo,
) map[string]*data.AccountBalanceHistory {
	accountsMap := make(map[string]*data.AccountBalanceHistory)
	for _, userAccount := range accounts {
		acc := &data.AccountBalanceHistory{
			Address:         userAccount.Address,
			Balance:         userAccount.Balance,
			Timestamp:       time.Duration(timestamp),
			Token:           userAccount.TokenName,
			TokenNonce:      userAccount.TokenNonce,
			IsSender:        userAccount.IsSender,
			IsSmartContract: userAccount.IsSmartContract,
			Identifier:      converters.ComputeTokenIdentifier(userAccount.TokenName, userAccount.TokenNonce),
		}
		keyInMap := fmt.Sprintf("%s-%s-%d", acc.Address, acc.Token, acc.TokenNonce)
		accountsMap[keyInMap] = acc
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

	esdtToken := &esdt.ESDigitalToken{}
	for _, tokenData := range accountESDT.Account.Tokens {
		if tokenData.Identifier == accountESDT.TokenIdentifier && tokenData.Nonce == accountESDT.NFTNonce {
			value, _ := big.NewInt(0).SetString(tokenData.Balance, 10)
			esdtToken = &esdt.ESDigitalToken{
				Value:         value,
				Properties:    []byte(tokenData.Properties),
				TokenMetaData: tokenData.MetaData,
			}
		}
	}

	if esdtToken.Value == nil {
		return big.NewInt(0), "", nil, nil
	}

	tokenMetaData := converters.PrepareTokenMetaData(ap.addressPubkeyConverter, esdtToken)

	return esdtToken.Value, hex.EncodeToString(esdtToken.Properties), tokenMetaData, nil
}

// PutTokenMedataDataInTokens will put the TokenMedata in provided tokens data
func (ap *accountsProcessor) PutTokenMedataDataInTokens(tokensData []*data.TokenInfo) {
	for _, tokenData := range tokensData {
		if tokenData.Data != nil || tokenData.Nonce == 0 {
			continue
		}

		tokenKey := computeTokenKey(tokenData.Token, tokenData.Nonce)
		metadata, errLoad := ap.loadMetadataFromSystemAccount(tokenKey)
		if errLoad != nil {
			log.Warn("cannot load token metadata",
				"token identifier ", tokenData.Identifier,
				"error", errLoad.Error())

			continue
		}

		tokenData.Data = converters.PrepareTokenMetaData(ap.addressPubkeyConverter, &esdt.ESDigitalToken{TokenMetaData: metadata})
	}
}

func (ap *accountsProcessor) loadMetadataFromSystemAccount(tokenKey []byte) (*esdt.MetaData, error) {
	systemAccount, err := ap.accountsDB.LoadAccount(vmcommon.SystemAccountAddress)
	if err != nil {
		return nil, err
	}

	userAccount, ok := systemAccount.(coreData.UserAccountHandler)
	if !ok {
		return nil, indexer.ErrCannotCastAccountHandlerToUserAccount
	}

	marshaledData, err := userAccount.RetrieveValueFromDataTrieTracker(tokenKey)
	if err != nil {
		return nil, err
	}

	esdtData := &esdt.ESDigitalToken{}
	err = ap.internalMarshalizer.Unmarshal(esdtData, marshaledData)
	if err != nil {
		return nil, err
	}

	return esdtData.TokenMetaData, nil
}

func computeTokenKey(token string, nonce uint64) []byte {
	tokenKey := []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier + token)
	if nonce > 0 {
		nonceBig := big.NewInt(0).SetUint64(nonce)
		tokenKey = append(tokenKey, nonceBig.Bytes()...)
	}

	return tokenKey
}
