package converters

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
)

const (
	ipfsURL            = "https://ipfs.io/ipfs/"
	ipfsNoSecurePrefix = "ipfs://"
	dwebPrefixURL      = "https://dweb.link/ipfs"

	pinataCloud = ".pinata.cloud/ipfs"
	secureURL   = "https://"
)

// PrepareTokenMetaData will prepare the token metadata in a friendly format for database
func PrepareTokenMetaData(tokenMetadata *outport.TokenMetaData) *data.TokenMetaData {
	if tokenMetadata == nil {
		return nil
	}

	return &data.TokenMetaData{
		Name:               tokenMetadata.Name,
		Creator:            tokenMetadata.Creator,
		Royalties:          tokenMetadata.Royalties,
		Hash:               tokenMetadata.Hash,
		URIs:               tokenMetadata.URIs,
		Attributes:         tokenMetadata.Attributes,
		Tags:               ExtractTagsFromAttributes(tokenMetadata.Attributes),
		MetaData:           ExtractMetaDataFromAttributes(tokenMetadata.Attributes),
		NonEmptyURIs:       nonEmptyURIs(tokenMetadata.URIs),
		WhiteListedStorage: whiteListedStorage(tokenMetadata.URIs),
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

func whiteListedStorage(uris [][]byte) bool {
	if len(uris) == 0 {
		return false
	}

	uri := string(uris[0])

	whiteListed := strings.HasPrefix(string(uris[0]), ipfsURL)
	whiteListed = whiteListed || strings.HasPrefix(uri, ipfsNoSecurePrefix)
	whiteListed = whiteListed || strings.HasPrefix(uri, dwebPrefixURL)
	whiteListed = whiteListed || (strings.Contains(uri, pinataCloud) && strings.HasPrefix(uri, secureURL))

	return whiteListed
}

// PrepareNFTUpdateData will prepare nfts update data
func PrepareNFTUpdateData(buffSlice *data.BufferSlice, updateNFTData []*data.NFTDataUpdate, accountESDT bool, index string) error {
	for _, nftUpdate := range updateNFTData {
		id := nftUpdate.Identifier
		if accountESDT {
			id = fmt.Sprintf("%s-%s", nftUpdate.Address, nftUpdate.Identifier)
		}

		metaData := []byte(fmt.Sprintf(`{"update":{ "_index":"%s","_id":"%s"}}%s`, index, id, "\n"))
		base64Attr := base64.StdEncoding.EncodeToString(nftUpdate.NewAttributes)

		newTags := ExtractTagsFromAttributes(nftUpdate.NewAttributes)
		newMetadata := ExtractMetaDataFromAttributes(nftUpdate.NewAttributes)

		marshalizedTags, errM := json.Marshal(newTags)
		if errM != nil {
			return errM
		}

		codeToExecute := `
			if (ctx._source.containsKey('data')) {
				ctx._source.data.attributes = params.attributes;
				if (!params.metadata.isEmpty() ) {
					ctx._source.data.metadata = params.metadata
				} else {
					if (ctx._source.data.containsKey('metadata')) {
						ctx._source.data.remove('metadata')
					}
				}
				if (params.tags != null) {
					ctx._source.data.tags = params.tags
				} else {
					if (ctx._source.data.containsKey('tags')) {
						ctx._source.data.remove('tags')
					}
				}
			}
`
		serializedData := []byte(fmt.Sprintf(`{"script": {"source": "%s","lang": "painless","params": {"attributes": "%s", "metadata": "%s", "tags": %s}}, "upsert": {}}`,
			FormatPainlessSource(codeToExecute), base64Attr, newMetadata, marshalizedTags),
		)
		if len(nftUpdate.URIsToAdd) != 0 {
			marshalizedURIS, err := json.Marshal(nftUpdate.URIsToAdd)
			if err != nil {
				return err
			}

			codeToExecute = `
				if (ctx._source.containsKey('data')) {
					if (!ctx._source.data.containsKey('uris')) {
						ctx._source.data.uris = params.uris;
					} else {
						int i;
						for ( i = 0; i < params.uris.length; i++) {
							boolean found = false;
							int j;
							for ( j = 0; j < ctx._source.data.uris.length; j++) {
								if ( params.uris.get(i) == ctx._source.data.uris.get(j) ) {
									found = true;
									break
								}
							}
							if ( !found ) {
								ctx._source.data.uris.add(params.uris.get(i))
							}
						}
					}
					ctx.nonEmptyURIs = true;
				}
`
			serializedData = []byte(fmt.Sprintf(`{"script": {"source": "%s","lang": "painless","params": {"uris": %s}},"upsert": {}}`, FormatPainlessSource(codeToExecute), marshalizedURIS))
		}

		err := buffSlice.PutData(metaData, serializedData)
		if err != nil {
			return err
		}
	}

	return nil
}
