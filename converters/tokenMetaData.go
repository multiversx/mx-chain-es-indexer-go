package converters

import (
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
)

// PrepareTokenMetaData will prepare the token metadata in a friendly format for database
func PrepareTokenMetaData(pubKeyConverter core.PubkeyConverter, esdtInfo *esdt.ESDigitalToken) *data.TokenMetaData {
	if check.IfNil(pubKeyConverter) {
		return nil
	}

	if esdtInfo == nil || esdtInfo.TokenMetaData == nil {
		return nil
	}

	creatorStr := ""
	if esdtInfo.TokenMetaData.Creator != nil {
		creatorStr = pubKeyConverter.Encode(esdtInfo.TokenMetaData.Creator)
	}

	return &data.TokenMetaData{
		Name:         string(esdtInfo.TokenMetaData.Name),
		Creator:      creatorStr,
		Royalties:    esdtInfo.TokenMetaData.Royalties,
		Hash:         esdtInfo.TokenMetaData.Hash,
		URIs:         esdtInfo.TokenMetaData.URIs,
		Attributes:   esdtInfo.TokenMetaData.Attributes,
		Tags:         ExtractTagsFromAttributes(esdtInfo.TokenMetaData.Attributes),
		MetaData:     ExtractMetaDataFromAttributes(esdtInfo.TokenMetaData.Attributes),
		NonEmptyURIs: nonEmptyURIs(esdtInfo.TokenMetaData.URIs),
	}
}

func nonEmptyURIs(uris [][]byte) bool {
	for _, uri := range uris {
		if len(uri) > 0 {
			return true
		}
	}

	return false
}

// PrepareNFTUpdateData will prepare nfts update data
func PrepareNFTUpdateData(buffSlice *data.BufferSlice, updateNFTData []*data.UpdateNFTData) error {
	for _, nftUpdate := range updateNFTData {
		id := fmt.Sprintf("%s-%s", nftUpdate.Address, nftUpdate.Identifier)
		metaData := []byte(fmt.Sprintf(`{"update":{"_id":"%s", "_type": "_doc"}}%s`, id, "\n"))
		serializedData := []byte(fmt.Sprintf(`{"script": {"source": "ctx._source.data.attributes = params.attributes","lang": "painless","params": {"attributes": "%s"}}}`, nftUpdate.NewAttributes))
		if len(nftUpdate.URIsToAdd) != 0 {
			serializedData = []byte(fmt.Sprintf(`{"script": {"source": "if (!ctx._source.data.containsKey('uris')) { ctx._source.data.uris = [ params.uris ]; } else {  ctx._source.data.uris.addAll(params.uris); }","lang": "painless","params": {"uris": %v}}}`, nftUpdate.URIsToAdd))
		}

		err := buffSlice.PutData(metaData, serializedData)
		if err != nil {
			return err
		}
	}

	return nil
}
