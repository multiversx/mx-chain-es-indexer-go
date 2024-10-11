package converters

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

const (
	ipfsURL            = "https://ipfs.io/ipfs/"
	ipfsNoSecurePrefix = "ipfs://"
	dwebPrefixURL      = "https://dweb.link/ipfs"

	pinataCloud = ".pinata.cloud/ipfs"
	secureURL   = "https://"
)

// PrepareTokenMetaData will prepare the token metadata in a friendly format for database
func PrepareTokenMetaData(tokenMetadata *alteredAccount.TokenMetaData) *data.TokenMetaData {
	if tokenMetadata == nil {
		return nil
	}

	var uris [][]byte
	for _, uri := range tokenMetadata.URIs {
		truncatedURI := TruncateFieldIfExceedsMaxLengthBase64(string(uri))
		uris = append(uris, []byte(truncatedURI))
	}

	tags := ExtractTagsFromAttributes(tokenMetadata.Attributes)
	attributes := TruncateFieldIfExceedsMaxLengthBase64(string(tokenMetadata.Attributes))
	return &data.TokenMetaData{
		Name:               TruncateFieldIfExceedsMaxLength(tokenMetadata.Name),
		Creator:            tokenMetadata.Creator,
		Royalties:          tokenMetadata.Royalties,
		Hash:               tokenMetadata.Hash,
		URIs:               uris,
		Attributes:         []byte(attributes),
		Tags:               TruncateSliceElementsIfExceedsMaxLength(tags),
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
func PrepareNFTUpdateData(buffSlice *data.BufferSlice, updateNFTData []*data.NFTDataUpdate, isAccountsESDTIndex bool, index string) error {
	for _, nftUpdate := range updateNFTData {
		id := nftUpdate.Identifier
		if isAccountsESDTIndex {
			id = fmt.Sprintf("%s-%s", nftUpdate.Address, nftUpdate.Identifier)
		}

		metaData := []byte(fmt.Sprintf(`{"update":{ "_index":"%s","_id":"%s"}}%s`, index, id, "\n"))
		freezeOrUnfreezeTokenIndex := (nftUpdate.Freeze || nftUpdate.UnFreeze) && !isAccountsESDTIndex
		if freezeOrUnfreezeTokenIndex {
			err := buffSlice.PutData(metaData, prepareSerializeDataForFreezeAndUnFreeze(nftUpdate))
			if err != nil {
				return err
			}
			continue
		}
		pauseOrUnPauseTokenIndex := (nftUpdate.Pause || nftUpdate.UnPause) && !isAccountsESDTIndex
		if pauseOrUnPauseTokenIndex {
			err := buffSlice.PutData(metaData, prepareSerializedDataForPauseAndUnPause(nftUpdate))
			if err != nil {
				return err
			}

			continue
		}
		if nftUpdate.NewMetaData != nil {
			serializedData, err := prepareSerializedDataForMetaDataRecreate(nftUpdate)
			if err != nil {
				return err
			}
			err = buffSlice.PutData(metaData, serializedData)
			if err != nil {
				return err
			}

			continue
		}

		if nftUpdate.NewCreator != "" {
			err := buffSlice.PutData(metaData, prepareSerializeDataForNewCreator(nftUpdate))
			if err != nil {
				return err
			}

			continue
		}
		if nftUpdate.NewRoyalties.HasValue {
			err := buffSlice.PutData(metaData, prepareSerializeDataForNewRoyalties(nftUpdate))
			if err != nil {
				return err
			}

			continue
		}

		truncatedAttributes := TruncateFieldIfExceedsMaxLengthBase64(string(nftUpdate.NewAttributes))
		base64Attr := base64.StdEncoding.EncodeToString([]byte(truncatedAttributes))
		newTags := TruncateSliceElementsIfExceedsMaxLength(ExtractTagsFromAttributes(nftUpdate.NewAttributes))
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
			uris := make([]string, 0, len(nftUpdate.URIsToAdd))
			for _, uri := range nftUpdate.URIsToAdd {
				uris = append(uris, base64.StdEncoding.EncodeToString(uri))
			}
			marshalizedURIS, err := json.Marshal(TruncateSliceElementsIfExceedsMaxLength(uris))
			if err != nil {
				return err
			}

			codeToExecute = `
				if (ctx._source.containsKey('data')) {
					if ((!ctx._source.data.containsKey('uris')) || (params.set)) {
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
					ctx._source.data.nonEmptyURIs = true;
				}
`
			serializedData = []byte(fmt.Sprintf(`{"script": {"source": "%s","lang": "painless","params": {"uris": %s, "set":%t}},"upsert": {}}`, FormatPainlessSource(codeToExecute), marshalizedURIS, nftUpdate.SetURIs))
		}

		err := buffSlice.PutData(metaData, serializedData)
		if err != nil {
			return err
		}
	}

	return nil
}

func prepareSerializeDataForFreezeAndUnFreeze(nftUpdateData *data.NFTDataUpdate) []byte {
	frozen := nftUpdateData.Freeze
	codeToExecute := `
			ctx._source.frozen = params.frozen
`
	serializedData := []byte(fmt.Sprintf(`{"script": {"source": "%s","lang": "painless","params": {"frozen": %t}}, "upsert": {}}`,
		FormatPainlessSource(codeToExecute), frozen),
	)

	return serializedData
}

func prepareSerializedDataForPauseAndUnPause(nftUpdateData *data.NFTDataUpdate) []byte {
	paused := nftUpdateData.Pause
	codeToExecute := `
			ctx._source.paused = params.paused
`
	serializedData := []byte(fmt.Sprintf(`{"script": {"source": "%s","lang": "painless","params": {"paused": %t}}, "upsert": {}}`,
		FormatPainlessSource(codeToExecute), paused),
	)

	return serializedData
}

func prepareSerializedDataForMetaDataRecreate(nftUpdateData *data.NFTDataUpdate) ([]byte, error) {
	tokenMetaDataBytes, err := json.Marshal(nftUpdateData.NewMetaData)
	if err != nil {
		return nil, err
	}

	codeToExecute := `
			ctx._source.data = params.metaData;
`
	serializedData := []byte(fmt.Sprintf(`{"script": {"source": "%s","lang": "painless","params": {"metaData": %s}}, "upsert": {}}`,
		FormatPainlessSource(codeToExecute), tokenMetaDataBytes),
	)

	return serializedData, nil
}

func prepareSerializeDataForNewRoyalties(nftUpdateData *data.NFTDataUpdate) []byte {
	codeToExecute := `
			if (ctx._source.containsKey('data')) {
				ctx._source.data.royalties = params.royalties;
			}
`
	serializedData := []byte(fmt.Sprintf(`{"script": {"source": "%s","lang": "painless","params": {"royalties": %d}}, "upsert": {}}`,
		FormatPainlessSource(codeToExecute), nftUpdateData.NewRoyalties.Value),
	)

	return serializedData
}

func prepareSerializeDataForNewCreator(nftUpdateData *data.NFTDataUpdate) []byte {
	codeToExecute := `
			if (ctx._source.containsKey('data')) {
				ctx._source.data.creator = params.creator;
			}
`
	serializedData := []byte(fmt.Sprintf(`{"script": {"source": "%s","lang": "painless","params": {"creator": "%s"}}, "upsert": {}}`,
		FormatPainlessSource(codeToExecute), nftUpdateData.NewCreator),
	)

	return serializedData
}
