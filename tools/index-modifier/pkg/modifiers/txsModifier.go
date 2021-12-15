package modifiers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ElrondNetwork/elastic-indexer-go/tools/index-modifier/pkg/modifiers/utils"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/process/transactions"
	"github.com/ElrondNetwork/elastic-indexer-go/process/transactions/datafield"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	factoryMarshalizer "github.com/ElrondNetwork/elrond-go-core/marshal/factory"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("index-modifier/pkg/alterindex")

type responseTransactionsBulk struct {
	Hits struct {
		Hits []struct {
			ID     string            `json:"_id"`
			Source *data.Transaction `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type txsModifier struct {
	pubKeyConverter     core.PubkeyConverter
	operationDataParser transactions.DataFieldParser
}

func NewTxsModifier() (*txsModifier, error) {
	pubKeyConverter, parser, err := createPubKeyConverterAndParser()
	if err != nil {
		return nil, err
	}

	return &txsModifier{
		pubKeyConverter:     pubKeyConverter,
		operationDataParser: parser,
	}, nil
}

func createOperationParser(pubkeyConverter core.PubkeyConverter) (transactions.DataFieldParser, error) {
	shardCoordinator, err := utils.NewMultiShardCoordinator(3, 0)
	if err != nil {
		return nil, err
	}
	marshalizer, err := factoryMarshalizer.NewMarshalizer(factoryMarshalizer.GogoProtobuf)

	arguments := &datafield.ArgsOperationDataFieldParser{
		PubKeyConverter:  pubkeyConverter,
		Marshalizer:      marshalizer,
		ShardCoordinator: shardCoordinator,
	}

	return datafield.NewOperationDataFieldParser(arguments)
}

func createPubKeyConverterAndParser() (core.PubkeyConverter, transactions.DataFieldParser, error) {
	pubKeyConverter, err := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	if err != nil {
		return nil, nil, err
	}

	parser, err := createOperationParser(pubKeyConverter)
	if err != nil {
		return nil, nil, err
	}

	return pubKeyConverter, parser, nil
}

func (tm *txsModifier) Modify(responseBody []byte) ([]*bytes.Buffer, error) {
	responseTxs := &responseTransactionsBulk{}
	err := json.Unmarshal(responseBody, responseTxs)
	if err != nil {
		return nil, err
	}

	buffSlice := data.NewBufferSlice()
	for _, hit := range responseTxs.Hits.Hits {
		if hit.Source.Sender == "4294967295" || hit.Source.Status == "pending" {
			continue
		}
		errPrep := tm.prepareTxForIndexing(hit.Source)
		if errPrep != nil {
			log.Warn("cannot prepare transaction",
				"error", errPrep.Error(),
				"hash", hit.ID,
			)
			continue
		}

		meta, serializedData, errSerialize := serializeTx(hit.ID, hit.Source)
		if errSerialize != nil {
			log.Warn("cannot serialize transaction",
				"error", errSerialize.Error(),
				"hash", hit.ID,
			)
			continue
		}

		errPut := buffSlice.PutData(meta, serializedData)
		if errPut != nil {
			log.Warn("cannot put transaction",
				"error", errPut.Error(),
				"hash", hit.ID,
			)
			continue
		}
	}

	return buffSlice.Buffers(), nil
}

func serializeTx(hash string, tx *data.Transaction) ([]byte, []byte, error) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, hash, "\n"))
	serializedData, errPrepareReceipt := json.Marshal(tx)
	if errPrepareReceipt != nil {
		return nil, nil, errPrepareReceipt
	}

	return meta, serializedData, nil
}

func (tm *txsModifier) prepareTxForIndexing(tx *data.Transaction) error {
	sndAddr, err := tm.pubKeyConverter.Decode(tx.Sender)
	if err != nil {
		return err
	}
	rcvAddr, err := tm.pubKeyConverter.Decode(tx.Receiver)
	if err != nil {
		return err
	}

	res := tm.operationDataParser.Parse(tx.Data, sndAddr, rcvAddr)

	tx.Operation = res.Operation
	tx.Function = res.Function
	tx.ESDTValues = res.ESDTValues
	tx.Tokens = res.Tokens
	tx.Receivers = res.Receivers
	tx.ReceiversShardIDs = res.ReceiversShardID

	return nil
}
