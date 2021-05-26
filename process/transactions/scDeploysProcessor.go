package transactions

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/parsers"
	"github.com/ElrondNetwork/elrond-go/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

const (
	delegationManagerAddress = "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqylllslmq6y6"
)

type scDeploysProc struct {
	argumentsParserExtended *argumentsParserExtended
	pubKeyConverter         core.PubkeyConverter
	scDeployReceiverAddr    string
	selfShardID             uint32
}

type deployerDto struct {
	receiver string
	sender   string
	nonce    uint64
	data     []byte
	txHash   string
}

func newScDeploysProc(pubKeyConverter core.PubkeyConverter, selfShardID uint32) *scDeploysProc {
	scDeployReceiver := make([]byte, 32)
	argsParser := parsers.NewCallArgsParser()

	return &scDeploysProc{
		selfShardID:             selfShardID,
		pubKeyConverter:         pubKeyConverter,
		scDeployReceiverAddr:    pubKeyConverter.Encode(scDeployReceiver),
		argumentsParserExtended: newArgumentsParser(argsParser),
	}
}

func (sc *scDeploysProc) searchSCDeployTransactionsOrSCRS(
	txs []*data.Transaction,
	scrs []*data.ScResult,
) []*data.ScDeployInfo {
	scDeploys := make([]*data.ScDeployInfo, 0)
	for _, tx := range txs {
		deployer := &deployerDto{sender: tx.Sender, receiver: tx.Receiver, nonce: tx.Nonce, txHash: tx.Hash, data: tx.Data}
		deployInfo, ok := sc.searchNormalDeploy(deployer, tx.SmartContractResults)
		if ok {
			scDeploys = append(scDeploys, deployInfo)
			continue
		}

		deployInfo, ok = sc.searchDelegationManagerDeploy(deployer, tx.SmartContractResults)
		if ok {
			scDeploys = append(scDeploys, deployInfo)
			continue
		}
	}

	deploysFromSCRSNormal, ok := sc.searchForSCRWithDeployNormal(scrs)
	if ok {
		scDeploys = append(scDeploys, deploysFromSCRSNormal...)
	}

	deploysFromSCRSNormalManager, ok := sc.searchForSCRWithDelegationManagerDeploy(scrs)
	if ok {
		scDeploys = append(scDeploys, deploysFromSCRSNormalManager...)
	}

	return scDeploys
}

func (sc *scDeploysProc) searchForSCRWithDelegationManagerDeploy(scrs []*data.ScResult) ([]*data.ScDeployInfo, bool) {
	if sc.selfShardID != core.MetachainShardId {
		return nil, false
	}

	scDeploys := make([]*data.ScDeployInfo, 0)
	for _, scr := range scrs {
		deployer := &deployerDto{sender: scr.Sender, receiver: scr.Receiver, nonce: scr.Nonce, txHash: scr.Hash, data: scr.Data}
		deployInfo, ok := sc.searchDelegationManagerDeploy(deployer, scrs)
		if !ok {
			continue
		}

		scDeploys = append(scDeploys, deployInfo)
	}

	if len(scDeploys) == 0 {
		return nil, false
	}

	return scDeploys, true
}

func (sc *scDeploysProc) searchForSCRWithDeployNormal(scrs []*data.ScResult) ([]*data.ScDeployInfo, bool) {
	scDeploys := make([]*data.ScDeployInfo, 0)
	for _, scr := range scrs {
		if scr.Receiver != sc.scDeployReceiverAddr {
			continue
		}
		deployer := &deployerDto{sender: scr.Sender, receiver: scr.Receiver, nonce: scr.Nonce, txHash: scr.Hash, data: scr.Data}
		deployInfo, ok := sc.searchNormalDeploy(deployer, scrs)
		if !ok {
			continue
		}

		scDeploys = append(scDeploys, deployInfo)
	}

	if len(scDeploys) == 0 {
		return nil, false
	}

	return scDeploys, true
}

func (sc *scDeploysProc) searchNormalDeploy(deployer *deployerDto, scrs []*data.ScResult) (*data.ScDeployInfo, bool) {
	if deployer.receiver != sc.scDeployReceiverAddr {
		return nil, false
	}

	if len(scrs) < 1 {
		return nil, false
	}

	for _, scr := range scrs {
		if !sc.argumentsParserExtended.hasOKPrefix(string(scr.Data)) ||
			scr.Sender != sc.scDeployReceiverAddr ||
			scr.Receiver != deployer.sender {
			continue
		}

		scAddress, err := sc.computeContractAddress(deployer.sender, deployer.nonce, factory.ArwenVirtualMachine)
		if err != nil {
			continue
		}

		return &data.ScDeployInfo{
			ScAddress: scAddress,
			TxHash:    deployer.txHash,
			Creator:   deployer.sender,
		}, true
	}

	return nil, false
}

func (sc *scDeploysProc) searchDelegationManagerDeploy(deployer *deployerDto, scrs []*data.ScResult) (*data.ScDeployInfo, bool) {
	if sc.selfShardID != core.MetachainShardId {
		return nil, false
	}

	if deployer.receiver != delegationManagerAddress {
		return nil, false
	}

	if !bytes.HasPrefix(deployer.data, []byte("createNewDelegationContract")) {
		return nil, false
	}

	if len(scrs) < 3 {
		return nil, false
	}

	for _, scr := range scrs {
		isSCRFromDeployAddress := scr.Sender == delegationManagerAddress && scr.Receiver == deployer.sender
		if !isSCRFromDeployAddress {
			continue
		}

		if !sc.argumentsParserExtended.hasOKPrefix(string(scr.Data)) {
			continue
		}

		splitData := sc.argumentsParserExtended.split(string(scr.Data))
		if len(splitData) < 3 {
			continue
		}

		hexDecoded, err := hex.DecodeString(splitData[2])
		if err != nil {
			continue
		}

		scAddress := sc.pubKeyConverter.Encode(hexDecoded)
		return &data.ScDeployInfo{
			ScAddress: scAddress,
			TxHash:    deployer.txHash,
			Creator:   deployer.sender,
		}, true
	}

	return nil, false
}

func (sc *scDeploysProc) computeContractAddress(creatorAddressBech string, creatorNonce uint64, vmType []byte) (string, error) {
	creatorAddress, err := sc.pubKeyConverter.Decode(creatorAddressBech)
	if err != nil {
		return "", err
	}

	base := hashFromAddressAndNonce(creatorAddress, creatorNonce)
	prefixMask := createPrefixMask(vmType)
	suffixMask := createSuffixMask(creatorAddress)

	copy(base[:core.NumInitCharactersForScAddress], prefixMask)
	copy(base[len(base)-core.ShardIdentiferLen:], suffixMask)

	return sc.pubKeyConverter.Encode(base), nil
}

func hashFromAddressAndNonce(creatorAddress []byte, creatorNonce uint64) []byte {
	buffNonce := make([]byte, 8)
	binary.LittleEndian.PutUint64(buffNonce, creatorNonce)
	adrAndNonce := append(creatorAddress, buffNonce...)
	scAddress := keccak.Keccak{}.Compute(string(adrAndNonce))

	return scAddress
}

func createPrefixMask(vmType []byte) []byte {
	prefixMask := make([]byte, core.NumInitCharactersForScAddress-core.VMTypeLen)
	prefixMask = append(prefixMask, vmType...)

	return prefixMask
}

func createSuffixMask(creatorAddress []byte) []byte {
	return creatorAddress[len(creatorAddress)-2:]
}
