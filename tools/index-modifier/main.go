package main

import (
	"fmt"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/tools/index-modifier/pkg/modifiers"
	"github.com/ElrondNetwork/elastic-indexer-go/tools/index-modifier/pkg/reindex"
)

const (
	scrollClientAddress = ""
	bulkClientAddress   = ""
)

func main() {
	indexModifier, err := reindex.CreateIndexModifier(scrollClientAddress, bulkClientAddress)
	if err != nil {
		panic("cannot create index modifier: " + err.Error())
	}

	txsModifier, err := modifiers.NewTxsModifier()
	if err != nil {
		panic("cannot create transactions modifier: " + err.Error())
	}

	err = indexModifier.AlterIndex(indexer.TransactionsIndex, txsModifier.Modify)
	if err != nil {
		panic("cannot modify index: " + err.Error())
	}

	fmt.Println("done")
}
