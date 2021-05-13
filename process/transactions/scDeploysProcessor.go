package transactions

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

const (
	delegationManagerAddress = "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqylllslmq6y6"
)

type scDeploysProc struct {
	pubKeyConverter      core.PubkeyConverter
	scDeployReceiverAddr string
	selfShardID          uint32
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

	return &scDeploysProc{
		selfShardID:          selfShardID,
		pubKeyConverter:      pubKeyConverter,
		scDeployReceiverAddr: pubKeyConverter.Encode(scDeployReceiver),
	}
}

func (sc *scDeploysProc) searchSCDeployTransactionsOrSCRS(
	txs []*data.Transaction,
	scrs []*data.ScResult,
) []*data.ScDeployInfo {
	scDeploys := make([]*data.ScDeployInfo, 0)
	for _, tx := range txs {
		dto := &deployerDto{sender: tx.Sender, receiver: tx.Receiver, nonce: tx.Nonce, txHash: tx.Hash, data: tx.Data}
		deployInfo, ok := sc.searchNormalDeploy(dto, tx.SmartContractResults)
		if ok {
			scDeploys = append(scDeploys, deployInfo)
			continue
		}

		deployInfo, ok = sc.searchDelegationManagerDeploy(dto, tx.SmartContractResults)
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
		dto := &deployerDto{sender: scr.Sender, receiver: scr.Receiver, nonce: scr.Nonce, txHash: scr.Hash, data: scr.Data}
		deployInfo, ok := sc.searchDelegationManagerDeploy(dto, scrs)
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
		dto := &deployerDto{sender: scr.Sender, receiver: scr.Receiver, nonce: scr.Nonce, txHash: scr.Hash, data: scr.Data}
		deployInfo, ok := sc.searchNormalDeploy(dto, scrs)
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

func (sc *scDeploysProc) searchNormalDeploy(dto *deployerDto, scrs []*data.ScResult) (*data.ScDeployInfo, bool) {
	if dto.receiver != sc.scDeployReceiverAddr {
		return nil, false
	}

	if len(scrs) < 1 {
		return nil, false
	}

	for _, scr := range scrs {
		if !bytes.HasPrefix(scr.Data, []byte("@6f6b")) ||
			scr.Sender != sc.scDeployReceiverAddr ||
			scr.Receiver != dto.sender {
			continue
		}

		scAddress, err := sc.computeContractAddress(dto.sender, dto.nonce, factory.ArwenVirtualMachine)
		if err != nil {
			continue
		}

		return &data.ScDeployInfo{
			ScAddress: scAddress,
			TxHash:    dto.txHash,
			Creator:   dto.sender,
		}, true
	}

	return nil, false
}

func (sc *scDeploysProc) searchDelegationManagerDeploy(dto *deployerDto, scrs []*data.ScResult) (*data.ScDeployInfo, bool) {
	if sc.selfShardID != core.MetachainShardId {
		return nil, false
	}

	if dto.receiver != delegationManagerAddress {
		return nil, false
	}

	if !bytes.HasPrefix(dto.data, []byte("createNewDelegationContract")) {
		return nil, false
	}

	if len(scrs) < 3 {
		return nil, false
	}

	for _, scr := range scrs {
		if !(scr.Sender == delegationManagerAddress && scr.Receiver == dto.sender) {
			continue
		}

		if !bytes.HasPrefix(scr.Data, []byte("@6f6b")) {
			continue
		}

		splitData := bytes.Split(scr.Data, []byte("@"))
		if len(splitData) < 3 {
			continue
		}

		hexDecoded, err := hex.DecodeString(string(splitData[2]))
		if err != nil {
			continue
		}

		scAddress := sc.pubKeyConverter.Encode(hexDecoded)
		return &data.ScDeployInfo{
			ScAddress: scAddress,
			TxHash:    dto.txHash,
			Creator:   dto.sender,
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
