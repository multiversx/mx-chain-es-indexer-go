package logstoevents

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/core/sharding"
	"github.com/multiversx/mx-chain-es-indexer-go/client"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/migrations/dtos"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/tidwall/gjson"
)

var log = logger.GetOrCreate("split-logs")

type ArgsEventsProc struct {
	SourceCluster      dtos.ClusterSettings
	DestinationCluster dtos.ClusterSettings
}

type eventsProcessor struct {
	sourceESClient      EsClient
	destinationESClient EsClient

	count            int
	addressConverter core.PubkeyConverter
}

func NewEventsProcessor(args ArgsEventsProc) (*eventsProcessor, error) {
	sourceClient, err := client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{args.SourceCluster.URL},
		Username:  args.SourceCluster.User,
		Password:  args.SourceCluster.Password,
	})
	if err != nil {
		return nil, err
	}
	destinationClient, err := client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{args.DestinationCluster.URL},
		Username:  args.DestinationCluster.User,
		Password:  args.DestinationCluster.Password,
	})

	pubKeyConverter, err := pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
	if err != nil {
		return nil, err
	}
	return &eventsProcessor{
		sourceESClient:      sourceClient,
		destinationESClient: destinationClient,
		addressConverter:    pubKeyConverter,
	}, nil
}

func (ep *eventsProcessor) SplitLogIndexInEvents(migrationID, sourceIndex, destinationIndex string) error {
	migrationInfo, err := ep.checkStatusOfSplit(migrationID)
	if err != nil {
		return err
	}

	lastTimestamp := migrationInfo.Timestamp

	done := false
	for !done {
		query := computeQueryBasedOnTimestamp(lastTimestamp)
		response := &dtos.ResponseLogsSearch{}
		err = ep.sourceESClient.DoSearchRequest(context.Background(), sourceIndex, bytes.NewBuffer(query), response)
		if err != nil {
			return err
		}

		lastTimestamp, done, err = ep.prepareAndIndexEventsFromLogs(migrationID, response, destinationIndex)
		if err != nil {
			return err
		}
		ep.count++

		log.Info("indexing events", "bulk-count", ep.count, "current-timestamp", lastTimestamp)

	}

	return nil
}

func (ep *eventsProcessor) prepareAndIndexEventsFromLogs(splitID string, logsResponse *dtos.ResponseLogsSearch, destinationIndex string) (uint64, bool, error) {
	logEvents := make([]*data.LogEvent, 0)
	for _, dbLog := range logsResponse.Hits.Hits {
		dbLogEvents, err := ep.createEventsFromLog(dbLog.ID, dbLog.Source)
		if err != nil {
			return 0, false, err
		}

		logEvents = append(logEvents, dbLogEvents...)
	}

	buffSlice := data.NewBufferSlice(0)
	for _, dbLog := range logEvents {
		err := serializeLogEvent(dbLog, buffSlice)
		if err != nil {
			return 0, false, err
		}
	}

	err := ep.doBulkRequests(destinationIndex, buffSlice.Buffers())
	if err != nil {
		return 0, false, err
	}

	// if we get less documents then the search size means the migration is done
	migrationDone := len(logsResponse.Hits.Hits) != searchSIZE

	lastEvent := logEvents[len(logEvents)-1]
	err = ep.writeStatusOfSplit(splitID, uint64(lastEvent.Timestamp), migrationDone)
	if err != nil {
		return 0, false, err
	}

	return uint64(lastEvent.Timestamp), migrationDone, nil
}

func (ep *eventsProcessor) doBulkRequests(index string, buffSlice []*bytes.Buffer) error {
	var err error
	for idx := range buffSlice {
		err = ep.destinationESClient.DoBulkRequest(context.Background(), buffSlice[idx], index)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ep *eventsProcessor) createEventsFromLog(txHash string, log *data.Logs) ([]*data.LogEvent, error) {
	logEvents := make([]*data.LogEvent, 0, len(log.Events))
	for _, event := range log.Events {
		addressBytes, err := ep.addressConverter.Decode(log.Address)
		if err != nil {
			return nil, err
		}

		eventShardID := sharding.ComputeShardID(addressBytes, 3)
		logEvents = append(logEvents, &data.LogEvent{
			ID:             fmt.Sprintf("%s-%d-%d", txHash, eventShardID, event.Order),
			TxHash:         txHash,
			OriginalTxHash: log.OriginalTxHash,
			LogAddress:     log.Address,
			Address:        event.Address,
			Identifier:     event.Identifier,
			Data:           hex.EncodeToString(event.Data),
			AdditionalData: hexEncodeSlice(event.AdditionalData),
			Topics:         hexEncodeSlice(event.Topics),
			Order:          event.Order,
			ShardID:        eventShardID,
			Timestamp:      log.Timestamp,
		})
	}

	return logEvents, nil
}

func (ep *eventsProcessor) writeStatusOfSplit(splitID string, timestamp uint64, isDone bool) error {
	status := dtos.MigrationInProgress
	if isDone {
		status = dtos.MigrationCompleted
	}

	migrationInfo := dtos.MigrationInfo{
		Status:    status,
		Timestamp: timestamp,
	}

	meta := []byte(fmt.Sprintf(`{ "index" : { "_index":"%s", "_id" : "%s" } }%s`, dataindexer.ValuesIndex, splitID, "\n"))
	migrationInfoBytes, err := json.Marshal(migrationInfo)
	if err != nil {
		return err
	}

	buffSlice := data.NewBufferSlice(0)
	err = buffSlice.PutData(meta, migrationInfoBytes)
	if err != nil {
		return err
	}

	return ep.destinationESClient.DoBulkRequest(context.Background(), buffSlice.Buffers()[0], "")
}

func (ep *eventsProcessor) checkStatusOfSplit(splitID string) (*dtos.MigrationInfo, error) {
	var response json.RawMessage
	err := ep.destinationESClient.DoMultiGet(context.Background(), []string{splitID}, dataindexer.ValuesIndex, true, &response)
	if err != nil {
		return nil, err
	}

	numOfDocs := gjson.Get(string(response), "docs.#").Int()
	if numOfDocs == 0 {
		return nil, nil
	}

	found := gjson.Get(string(response), "docs.0.found")
	if !found.Bool() {
		return &dtos.MigrationInfo{}, nil
	}

	responseBytes := []byte(gjson.Get(string(response), "docs.0._source").String())

	migrationInfo := &dtos.MigrationInfo{}
	err = json.Unmarshal(responseBytes, migrationInfo)
	return migrationInfo, err
}

func hexEncodeSlice(input [][]byte) []string {
	hexEncoded := make([]string, 0, len(input))
	for idx := 0; idx < len(input); idx++ {
		hexEncoded = append(hexEncoded, hex.EncodeToString(input[idx]))
	}
	if len(hexEncoded) == 0 {
		return nil
	}

	return hexEncoded
}
