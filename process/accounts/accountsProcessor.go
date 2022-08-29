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
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("indexer/process/accounts")

// accountsProcessor is a structure responsible for processing accounts
type accountsProcessor struct {
	addressPubkeyConverter core.PubkeyConverter
	balanceConverter       indexer.BalanceConverter
	shardID                uint32
}

// NewAccountsProcessor will create a new instance of accounts processor
func NewAccountsProcessor(
	addressPubkeyConverter core.PubkeyConverter,
	balanceConverter indexer.BalanceConverter,
	shardID uint32,
) (*accountsProcessor, error) {
	if check.IfNil(addressPubkeyConverter) {
		return nil, indexer.ErrNilPubkeyConverter
	}
	if check.IfNil(balanceConverter) {
		return nil, indexer.ErrNilBalanceConverter
	}

	return &accountsProcessor{
		addressPubkeyConverter: addressPubkeyConverter,
		balanceConverter:       balanceConverter,
		shardID:                shardID,
	}, nil
}

// TODO: refactor this as the altered accounts are already computed on the node. EN-12389
// GetAccounts will get accounts for regular operations and esdt operations
func (ap *accountsProcessor) GetAccounts(alteredAccounts data.AlteredAccountsHandler, coreAlteredAccounts map[string]*outport.AlteredAccount) ([]*data.Account, []*data.AccountESDT) {
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
	account *outport.AlteredAccount,
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
				IsNFTCreate:     info.IsNFTCreate,
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
func (ap *accountsProcessor) PrepareRegularAccountsMap(timestamp uint64, accounts []*data.Account) map[string]*data.AccountInfo {
	accountsMap := make(map[string]*data.AccountInfo)
	for _, userAccount := range accounts {
		address := userAccount.UserAccount.Address
		addressBytes, err := ap.addressPubkeyConverter.Decode(address)
		if err != nil {
			log.Warn("accountsProcessor.PrepareRegularAccountsMap: cannot decode address", "address", address, "error", err)
			continue
		}
		balance, ok := big.NewInt(0).SetString(userAccount.UserAccount.Balance, 10)
		if !ok {
			log.Warn("accountsProcessor.PrepareRegularAccountsMap: cannot cast account's balance to big int", "value", userAccount.UserAccount.Balance)
			continue
		}

		balanceAsFloat := ap.balanceConverter.ComputeBalanceAsFloat(balance)
		acc := &data.AccountInfo{
			Address:                  address,
			Nonce:                    userAccount.UserAccount.Nonce,
			Balance:                  converters.BigIntToString(balance),
			BalanceNum:               balanceAsFloat,
			IsSender:                 userAccount.IsSender,
			IsSmartContract:          core.IsSmartContractAddress(addressBytes),
			TotalBalanceWithStake:    converters.BigIntToString(balance),
			TotalBalanceWithStakeNum: balanceAsFloat,
			Timestamp:                time.Duration(timestamp),
			ShardID:                  ap.shardID,
		}

		accountsMap[address] = acc
	}

	return accountsMap
}

// PrepareAccountsMapESDT will prepare a map of accounts with ESDT tokens
func (ap *accountsProcessor) PrepareAccountsMapESDT(
	timestamp uint64,
	accounts []*data.AccountESDT,
	tagsCount data.CountTags,
) (map[string]*data.AccountInfo, data.TokensHandler) {
	tokensData := data.NewTokensInfo()
	accountsESDTMap := make(map[string]*data.AccountInfo)
	for _, accountESDT := range accounts {
		address := accountESDT.Account.Address
		addressBytes, err := ap.addressPubkeyConverter.Decode(address)
		if err != nil {
			log.Warn("accountsProcessor.PrepareAccountsMapESDT: cannot decode address", "address", address, "error", err)
			continue
		}
		balance, properties, tokenMetaData, err := ap.getESDTInfo(accountESDT)
		if err != nil {
			log.Warn("accountsProcessor.PrepareAccountsMapESDT: cannot get esdt info from account",
				"address", address,
				"error", err.Error())
			continue
		}

		if tokenMetaData != nil && accountESDT.IsNFTCreate {
			tagsCount.ParseTags(tokenMetaData.Tags)
		}

		tokenIdentifier := converters.ComputeTokenIdentifier(accountESDT.TokenIdentifier, accountESDT.NFTNonce)
		acc := &data.AccountInfo{
			Address:         address,
			TokenName:       accountESDT.TokenIdentifier,
			TokenIdentifier: tokenIdentifier,
			TokenNonce:      accountESDT.NFTNonce,
			Balance:         balance.String(),
			BalanceNum:      ap.balanceConverter.ComputeESDTBalanceAsFloat(balance),
			Properties:      properties,
			IsSender:        accountESDT.IsSender,
			IsSmartContract: core.IsSmartContractAddress(addressBytes),
			Data:            tokenMetaData,
			Timestamp:       time.Duration(timestamp),
			ShardID:         ap.shardID,
		}

		if acc.TokenNonce == 0 {
			acc.Type = core.FungibleESDT
		}

		keyInMap := fmt.Sprintf("%s-%s-%d", acc.Address, acc.TokenName, accountESDT.NFTNonce)
		accountsESDTMap[keyInMap] = acc

		if acc.Balance == "0" || acc.Balance == "" {
			continue
		}

		tokensData.Add(&data.TokenInfo{
			Token:      accountESDT.TokenIdentifier,
			Identifier: tokenIdentifier,
		})
	}

	return accountsESDTMap, tokensData
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
			ShardID:         ap.shardID,
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
func (ap *accountsProcessor) PutTokenMedataDataInTokens(tokensData []*data.TokenInfo, coreAlteredAccounts map[string]*outport.AlteredAccount) {
	for _, tokenData := range tokensData {
		if tokenData.Data != nil || tokenData.Nonce == 0 {
			continue
		}

		metadata, errLoad := ap.loadMetadataForToken(tokenData, coreAlteredAccounts)
		if errLoad != nil {
			log.Warn("accountsProcessor.PutTokenMedataDataInTokens: cannot load token metadata",
				"token identifier ", tokenData.Identifier,
				"error", errLoad.Error())

			continue
		}

		tokenData.Data = converters.PrepareTokenMetaData(ap.addressPubkeyConverter, &esdt.ESDigitalToken{TokenMetaData: metadata})
	}
}

func (ap *accountsProcessor) loadMetadataForToken(
	tokenData *data.TokenInfo,
	coreAlteredAccounts map[string]*outport.AlteredAccount,
) (*esdt.MetaData, error) {
	for _, account := range coreAlteredAccounts {
		for _, token := range account.Tokens {
			if tokenData.Token == token.Identifier && tokenData.Nonce == token.Nonce {
				return token.MetaData, nil
			}
		}
	}

	return nil, fmt.Errorf("%w for identifier %s and nonce %d", errTokenNotFound, tokenData.Identifier, tokenData.Nonce)
}
