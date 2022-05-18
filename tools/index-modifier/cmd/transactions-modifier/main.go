package main

import (
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/tools/index-modifier/pkg/alterindex"
	"github.com/ElrondNetwork/elastic-indexer-go/tools/index-modifier/pkg/modifiers"
)

const (
	scrollClientAddress = ""
	bulkClientAddress   = ""
)

func main() {
	indexModifier, err := alterindex.CreateIndexModifier(scrollClientAddress, bulkClientAddress)
	if err != nil {
		panic("cannot create index modifier: " + err.Error())
	}

	txsModifier, err := modifiers.NewTxsModifier()
	if err != nil {
		panic("cannot create transactions modifier: " + err.Error())
	}

	err = indexModifier.AlterIndex("transactions", "transactions", txsModifier.Modify)
	if err != nil {
		panic("cannot modify index: " + err.Error())
	}

	fmt.Println("done")
}
