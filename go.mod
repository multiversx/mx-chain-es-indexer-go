module github.com/ElrondNetwork/elastic-indexer-go

go 1.15

require (
	github.com/ElrondNetwork/elrond-go v1.2.5-0.20210720142639-fbb2186b264d
	github.com/ElrondNetwork/elrond-go-core v1.0.0
	github.com/ElrondNetwork/elrond-go-logger v1.0.4
	github.com/ElrondNetwork/elrond-vm-common v1.1.0
	github.com/elastic/go-elasticsearch/v7 v7.12.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_2 v1.2.26 => github.com/ElrondNetwork/arwen-wasm-vm v1.2.26

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_3 v1.3.25 => github.com/ElrondNetwork/arwen-wasm-vm v1.3.25
