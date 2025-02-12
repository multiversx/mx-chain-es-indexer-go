IMAGE_NAME=elastic-container
MAIN_CHAIN_IMAGE_NAME=main-chain-elastic-container
DEFAULT_ES_VERSION=7.16.2
PROMETHEUS_CONTAINER_NAME=prometheus_container
GRAFANA_CONTAINER_NAME=grafana_container
GRAFANA_VERSION=10.0.3
PROMETHEUS_VERSION=v2.46.0
INDICES_LIST=("rating" "transactions" "blocks" "validators" "miniblocks" "rounds" "accounts" "accountshistory" "receipts" "scresults" "accountsesdt" "accountsesdthistory" "epochinfo" "scdeploys" "tokens" "tags" "logs" "delegators" "operations" "esdts" "values" "events")


start() {
  ES_VERSION=$1
  if [ -z "${ES_VERSION}" ]; then
    ES_VERSION=${DEFAULT_ES_VERSION}
  fi

  docker pull docker.elastic.co/elasticsearch/elasticsearch:${ES_VERSION}

  docker rm ${IMAGE_NAME} 2> /dev/null
  docker rm ${MAIN_CHAIN_IMAGE_NAME} 2> /dev/null
  docker run -d --name "${IMAGE_NAME}" -p 9200:9200  -p 9300:9300 \
   -e "discovery.type=single-node" -e "xpack.security.enabled=false" -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" \
    docker.elastic.co/elasticsearch/elasticsearch:${ES_VERSION}
  docker run -d --name "${MAIN_CHAIN_IMAGE_NAME}" -p 9201:9200  -p 9301:9300 \
   -e "discovery.type=single-node" -e "xpack.security.enabled=false" -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" \
    docker.elastic.co/elasticsearch/elasticsearch:${ES_VERSION}

  # Wait elastic cluster to start
  echo "Waiting Elasticsearch cluster to start..."
  sleep 30s
}

stop() {
  docker stop "${IMAGE_NAME}"
  docker stop "${MAIN_CHAIN_IMAGE_NAME}"
}

delete() {
   for str in ${INDICES_LIST[@]}; do
      curl -XDELETE http://localhost:9200/$str-000001
      echo
   done

  curl -XDELETE http://localhost:9200/_template/*
  echo
}


IMAGE_OPEN_SEARCH=open-container
MAIN_CHAIN_IMAGE_OPEN_SEARCH=main-chain-open-container
DEFAULT_OPEN_SEARCH_VERSION=1.2.4

start_open_search() {
  OPEN_VERSION=$1
  if [ -z "${OPEN_VERSION}" ]; then
    OPEN_VERSION=${DEFAULT_OPEN_SEARCH_VERSION}
  fi

  docker pull opensearchproject/opensearch:${OPEN_VERSION}

  docker rm ${IMAGE_OPEN_SEARCH} 2> /dev/null
  docker rm ${MAIN_CHAIN_IMAGE_OPEN_SEARCH} 2> /dev/null
  docker run -d --name "${IMAGE_OPEN_SEARCH}" -p 9200:9200 -p 9600:9600 \
   -e "discovery.type=single-node" -e "plugins.security.disabled=true" -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" \
   opensearchproject/opensearch:${OPEN_VERSION}
  docker run -d --name "${MAIN_CHAIN_IMAGE_OPEN_SEARCH}" -p 9201:9200 -p 9601:9600 \
   -e "discovery.type=single-node" -e "plugins.security.disabled=true" -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" \
   opensearchproject/opensearch:${OPEN_VERSION}

}

stop_open_search() {
  docker stop "${IMAGE_OPEN_SEARCH}"
}

start_prometheus_and_grafana() {
 docker rm ${PROMETHEUS_CONTAINER_NAME} 2> /dev/null
 docker rm ${GRAFANA_CONTAINER_NAME} 2> /dev/null

 PROMETHEUS_CONFIG_FOLDER=$(pwd)/prometheus
 docker run --network="host" --name "${PROMETHEUS_CONTAINER_NAME}" -d -p 9090:9090 -v "${PROMETHEUS_CONFIG_FOLDER}/prometheus.yml":/etc/prometheus/prometheus.yml prom/prometheus:${PROMETHEUS_VERSION}
 docker run --network="host" --name "${GRAFANA_CONTAINER_NAME}" -d -p 3000:3000  grafana/grafana:${GRAFANA_VERSION}
}

stop_prometheus_and_grafana() {
  docker stop "${PROMETHEUS_CONTAINER_NAME}"
  docker stop "${GRAFANA_CONTAINER_NAME}"
}

"$@"
