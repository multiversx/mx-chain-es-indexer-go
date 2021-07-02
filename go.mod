module github.com/ElrondNetwork/elastic-indexer-go

go 1.15

require (
	github.com/ElrondNetwork/elrond-go v1.2.4-0.20210701140353-fb591fb0f9ae
	github.com/ElrondNetwork/elrond-go-logger v1.0.4
	github.com/ElrondNetwork/elrond-vm-common v0.3.4-0.20210625150635-525789454b6f
	github.com/elastic/go-elasticsearch/v7 v7.12.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_3 v1.3.19 => github.com/ElrondNetwork/arwen-wasm-vm v1.3.20-0.20210625154858-3c33a246ee32
