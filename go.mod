module github.com/ElrondNetwork/elastic-indexer-go

go 1.15

require (
	github.com/ElrondNetwork/elrond-go v1.2.5-0.20210721152414-e6091af3027a
	github.com/ElrondNetwork/elrond-go-core v1.0.1-0.20210721121720-f02fb03b2e1a
	github.com/ElrondNetwork/elrond-go-logger v1.0.5
	github.com/ElrondNetwork/elrond-vm-common v1.1.1-0.20210721110111-51b198fb52f4
	github.com/elastic/go-elasticsearch/v7 v7.12.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_2 v1.2.27 => github.com/ElrondNetwork/arwen-wasm-vm v1.2.27

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_3 v1.3.26 => github.com/ElrondNetwork/arwen-wasm-vm v1.3.26

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_4 v1.4.2 => github.com/ElrondNetwork/arwen-wasm-vm v1.4.2
