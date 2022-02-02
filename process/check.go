package process

import elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"

func checkArguments(arguments *ArgElasticProcessor) error {
	if arguments == nil {
		return elasticIndexer.ErrNilElasticProcessorArguments
	}
	if arguments.EnabledIndexes == nil {
		return elasticIndexer.ErrNilEnabledIndexesMap
	}
	if arguments.DBClient == nil {
		return elasticIndexer.ErrNilDatabaseClient
	}
	if arguments.StatisticsProc == nil {
		return elasticIndexer.ErrNilStatisticHandler
	}
	if arguments.BlockProc == nil {
		return elasticIndexer.ErrNilBlockHandler
	}
	if arguments.AccountsProc == nil {
		return elasticIndexer.ErrNilAccountsHandler
	}
	if arguments.MiniblocksProc == nil {
		return elasticIndexer.ErrNilMiniblocksHandler
	}
	if arguments.ValidatorsProc == nil {
		return elasticIndexer.ErrNilValidatorsHandler
	}
	if arguments.TransactionsProc == nil {
		return elasticIndexer.ErrNilTransactionsHandler
	}
	if arguments.LogsAndEventsProc == nil {
		return elasticIndexer.ErrNilLogsAndEventsHandler
	}
	if arguments.OperationsProc == nil {
		return elasticIndexer.ErrNilOperationsHandler
	}

	return nil
}
