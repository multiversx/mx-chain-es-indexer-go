package converters

import "github.com/ElrondNetwork/elastic-indexer-go/data"

// MergeAccountsInfoMaps will merge the provided accounts map in a new map
func MergeAccountsInfoMaps(
	accounts1 map[string]*data.AccountInfo,
	accounts2 map[string]*data.AccountInfo,
) map[string]*data.AccountInfo {
	resMap := make(map[string]*data.AccountInfo, len(accounts1)+len(accounts2))

	for key, value := range accounts1 {
		resMap[key] = value
	}

	for key, value := range accounts2 {
		resMap[key] = value
	}

	return resMap
}
