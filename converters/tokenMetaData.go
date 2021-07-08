package converters

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-vm-common/data/esdt"
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

	uris := esdtInfo.TokenMetaData.URIs
	if len(uris) == 0 {
		uris = nil
	}

	return &data.TokenMetaData{
		Name:       string(esdtInfo.TokenMetaData.Name),
		Creator:    creatorStr,
		Royalties:  esdtInfo.TokenMetaData.Royalties,
		Hash:       esdtInfo.TokenMetaData.Hash,
		URIs:       uris,
		Attributes: esdtInfo.TokenMetaData.Attributes,
		Tags:       ExtractTagsFromAttributes(esdtInfo.TokenMetaData.Attributes),
		MetaData:   ExtractMetaDataFromAttributes(esdtInfo.TokenMetaData.Attributes),
	}
}
