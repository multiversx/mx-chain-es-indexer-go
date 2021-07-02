module github.com/ElrondNetwork/elastic-indexer-go

go 1.15

require (
	github.com/ElrondNetwork/elrond-go v1.2.2-0.20210618111845-40305f3849b5
	github.com/ElrondNetwork/elrond-go-logger v1.0.4
	github.com/elastic/go-elasticsearch/v7 v7.12.0
	github.com/ElrondNetwork/elrond-vm-common v0.3.4-0.20210625102705-01a81e6a1fa4
	github.com/stretchr/testify v1.7.0
)

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_3 v1.3.19 => github.com/ElrondNetwork/arwen-wasm-vm v1.3.19
