module github.com/ElrondNetwork/elastic-indexer-go

go 1.15

require (
	github.com/ElrondNetwork/elrond-go v1.2.2-0.20210610132636-1422e5f4d81d
	github.com/ElrondNetwork/elrond-go-logger v1.0.4
	github.com/elastic/go-elasticsearch/v7 v7.12.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_3 v1.3.16 => github.com/ElrondNetwork/arwen-wasm-vm v1.3.16
