package main

import (
	"fmt"
	"github.com/ElrondNetwork/elastic-indexer-go/tools/index-modifier/pkg/modifiers"
	"github.com/ElrondNetwork/elastic-indexer-go/tools/index-modifier/pkg/reindex"
)

const (
	scrollClientAddress = "https://new-index.elrond.com"
	bulkClientAddress   = "http://localhost:9200"
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

	err = indexModifier.AlterIndex("transactions", txsModifier.Modify)
	if err != nil {
		panic("cannot modify index: " + err.Error())
	}

	fmt.Println("done")
}
