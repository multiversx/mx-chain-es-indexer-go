package collections

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
)

// ExtractAndSerializeCollectionsData will extra the accounts with NFT/SFT and serialize
func ExtractAndSerializeCollectionsData(
	accountsESDT map[string]*data.AccountInfo,
	buffSlice *data.BufferSlice,
	index string,
) error {
	for _, acct := range accountsESDT {
		shouldIgnore := acct.Type != core.NonFungibleESDT && acct.Type != core.SemiFungibleESDT
		if shouldIgnore {
			if acct.Balance != "0" || acct.TokenNonce == 0 {
				continue
			}
		}

		nonceBig := big.NewInt(0).SetUint64(acct.TokenNonce)
		hexEncodedNonce := hex.EncodeToString(nonceBig.Bytes())

		meta := []byte(fmt.Sprintf(`{ "update" : {"_index":"%s", "_id" : "%s" } }%s`, index, acct.Address, "\n"))
		codeToExecute := `
			if (('create' == ctx.op) && ('0' == params.value)) {
				ctx.op = 'noop';
			} else if ('0' == params.value) {
				if (!ctx._source.containsKey(params.col)) {
					ctx._source[params.col] = new HashMap();
				}
				ctx._source[params.col][params.nonce] = params.value
			} else {
				if (ctx._source.containsKey(params.col)) {
					ctx._source[params.col].remove(params.nonce);
					if (ctx._source[params.col].size() == 0) {
						ctx._source.remove(params.col)
					}
					if (ctx._source.size() == 0) {
						ctx.op = 'delete';
					}
				}
			}
`

		collection := fmt.Sprintf(`{"%s":{"%s": "%s"}}`, acct.TokenName, hexEncodedNonce, acct.Balance)
		serializedDataStr := fmt.Sprintf(`{"scripted_upsert": true, "script": {`+
			`"source": "%s",`+
			`"lang": "painless",`+
			`"params": { "col": "%s", "nonce": "%s", "value": "%s"}},`+
			`"upsert": %s}`,
			converters.FormatPainlessSource(codeToExecute), acct.TokenName, hexEncodedNonce, acct.Balance, collection)

		err := buffSlice.PutData(meta, []byte(serializedDataStr))
		if err != nil {
			return err
		}
	}

	return nil
}
