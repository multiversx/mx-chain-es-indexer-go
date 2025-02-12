TESTS_TO_RUN := $(shell go list ./... | grep -v integrationtests | grep -v mock)


test:
	@echo "  >  Running unit tests"
	go test -cover -race -coverprofile=coverage.txt -covermode=atomic -v ${TESTS_TO_RUN}

integration-tests:
	@echo " > Running integration tests"
	cd scripts && /bin/bash script.sh start ${ES_VERSION}
	go test -v ./integrationtests -tags integrationtests
	cd scripts && /bin/bash script.sh delete
	cd scripts && /bin/bash script.sh stop

long-tests:
	@-$(MAKE) delete-cluster-data
	go test -v ./integrationtests -tags integrationtests

start-cluster-with-kibana:
	@echo " > Starting Elasticsearch node and Kibana"
	docker-compose up -d

stop-cluster:
	docker-compose down

delete-cluster-data:
	cd scripts && /bin/bash script.sh delete

integration-tests-open-search:
	@echo " > Running integration tests open search"
	cd scripts && /bin/bash script.sh start_open_search ${OPEN_VERSION}
	go test -v ./integrationtests -tags integrationtests
	cd scripts && /bin/bash script.sh delete
	cd scripts && /bin/bash script.sh stop_open_search

INDEXER_IMAGE_NAME="elasticindexer"
INDEXER_IMAGE_TAG="latest"
DOCKER_FILE=Dockerfile
SOVEREIGN_DOCKER_FILE=Dockerfile-sovereign

docker-build:
	docker build \
		 -t ${INDEXER_IMAGE_NAME}:${INDEXER_IMAGE_TAG} \
		 -f ${DOCKER_FILE} \
		 .

docker-sovereign-build:
	docker build \
		 -t ${INDEXER_IMAGE_NAME}:${INDEXER_IMAGE_TAG} \
		 -f ${SOVEREIGN_DOCKER_FILE} \
		 .
