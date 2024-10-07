package innerTxs

import (
	"encoding/json"
	"fmt"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
)

// SerializeInnerTxs will serialize the provided array of inner transaction and add them in the BufferSlice
func (ip *innerTxsProcessor) SerializeInnerTxs(innerTxs []*data.InnerTransaction, buffSlice *data.BufferSlice, index string) error {
	for _, innerTx := range innerTxs {
		meta, serializedData, err := prepareSerializedDataForAInnerTx(innerTx, index)
		if err != nil {
			return err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return err
		}
	}

	return nil
}

func prepareSerializedDataForAInnerTx(
	innerTx *data.InnerTransaction,
	index string,
) ([]byte, []byte, error) {
	innerTxHash := innerTx.Hash
	innerTx.Hash = ""

	marshaledSCR, err := json.Marshal(innerTx)
	if err != nil {
		return nil, nil, err
	}

	meta := []byte(fmt.Sprintf(`{ "index" : { "_index":"%s","_id" : "%s" } }%s`, index, converters.JsonEscape(innerTxHash), "\n"))

	return meta, marshaledSCR, nil
}
