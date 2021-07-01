package converters

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
)

// PrepareTokenMetaData will prepare the token metadata in a friendly format for database
func PrepareTokenMetaData(pubKeyConverter core.PubkeyConverter, esdtInfo *esdt.ESDigitalToken) *data.TokenMetaData {
	if esdtInfo.TokenMetaData == nil {
		return nil
	}

	creatorStr := ""
	if esdtInfo.TokenMetaData.Creator != nil {
		creatorStr = pubKeyConverter.Encode(esdtInfo.TokenMetaData.Creator)
	}

	return &data.TokenMetaData{
		Name:       string(esdtInfo.TokenMetaData.Name),
		Creator:    creatorStr,
		Royalties:  esdtInfo.TokenMetaData.Royalties,
		Hash:       esdtInfo.TokenMetaData.Hash,
		URIs:       esdtInfo.TokenMetaData.URIs,
		Attributes: data.NewAttributesDTO(esdtInfo.TokenMetaData.Attributes),
	}
}
