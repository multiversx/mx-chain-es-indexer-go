module github.com/ElrondNetwork/elastic-indexer-go

go 1.15

require (
	github.com/ElrondNetwork/elrond-go v1.1.28-0.20210303132525-293022dea9c7
	github.com/ElrondNetwork/elrond-go-logger v1.0.4
	github.com/elastic/go-elasticsearch/v7 v7.10.0
	github.com/stretchr/testify v1.7.0
)

replace (
	github.com/ElrondNetwork/elrond-go v1.1.28-0.20210303132525-293022dea9c7 => /home/mihai/go/src/github.com/ElrondNetwork/elrond-go
)