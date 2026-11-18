IMAGE_NAME=elastic-container
DEFAULT_ES_VERSION=7.16.2
PROMETHEUS_CONTAINER_NAME=prometheus_container
GRAFANA_CONTAINER_NAME=grafana_container
GRAFANA_VERSION=10.0.3
PROMETHEUS_VERSION=v2.46.0
INDICES_LIST=("rating" "transactions" "blocks" "validators" "miniblocks" "rounds" "accounts" "accountshistory" "receipts" "scresults" "accountsesdt" "accountsesdthistory" "epochinfo" "scdeploys" "tokens" "tags" "logs" "delegators" "operations" "esdts" "values" "events" "executionresults")


start() {
  ES_VERSION=$1
  if [ -z "${ES_VERSION}" ]; then
    ES_VERSION=${DEFAULT_ES_VERSION}
  fi

  docker pull docker.elastic.co/elasticsearch/elasticsearch:${ES_VERSION}

  docker rm ${IMAGE_NAME} 2> /dev/null
  docker run -d --name "${IMAGE_NAME}" -p 9200:9200  -p 9300:9300 \
   -e "discovery.type=single-node" -e "xpack.security.enabled=false" -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" \
    docker.elastic.co/elasticsearch/elasticsearch:${ES_VERSION}
  # Wait elastic cluster to start
  echo "Waiting Elasticsearch cluster to start..."
  sleep 30s
  docker ps -a
}

stop() {
  docker stop "${IMAGE_NAME}"
}

delete() {
   for str in ${INDICES_LIST[@]}; do
      curl -XDELETE http://localhost:9200/$str-000001
      curl -s -o /dev/null -w "%{http_code}" -X GET localhost:9200/_ilm/policy/$str-policy | grep -q 200 && curl -X DELETE localhost:9200/_ilm/policy/$str-policy
      echo
   done

  curl -XDELETE http://localhost:9200/_template/*
  echo
}


IMAGE_OPEN_SEARCH=open-container
DEFAULT_OPEN_SEARCH_VERSION=1.2.4

start_open_search() {
  OPEN_VERSION=$1
  if [ -z "${OPEN_VERSION}" ]; then
    OPEN_VERSION=${DEFAULT_OPEN_SEARCH_VERSION}
  fi

  docker pull opensearchproject/opensearch:${OPEN_VERSION}

  docker rm ${IMAGE_OPEN_SEARCH} 2> /dev/null
  docker run -d --name "${IMAGE_OPEN_SEARCH}" -p 9200:9200 -p 9600:9600 \
   -e "discovery.type=single-node" -e "plugins.security.disabled=true" -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" \
   opensearchproject/opensearch:${OPEN_VERSION}

}

stop_open_search() {
  docker stop "${IMAGE_OPEN_SEARCH}"
}

RABBITMQ_CONTAINER_NAME=rabbitmq-container
DEFAULT_RABBITMQ_VERSION=3.13

start_rabbitmq() {
  RABBITMQ_VERSION=$1
  if [ -z "${RABBITMQ_VERSION}" ]; then
    RABBITMQ_VERSION=${DEFAULT_RABBITMQ_VERSION}
  fi

  docker pull rabbitmq:${RABBITMQ_VERSION}

  docker rm ${RABBITMQ_CONTAINER_NAME} 2> /dev/null
  docker run -d --name "${RABBITMQ_CONTAINER_NAME}" -p 5672:5672 rabbitmq:${RABBITMQ_VERSION}

  # Wait the RabbitMQ server to accept connections.
  # The check has to run with the rabbitmq user: if it runs with the root user before the server is started,
  # the erlang cookie file is created by root and the server cannot start
  echo "Waiting RabbitMQ server to start..."
  for _ in $(seq 1 60); do
    if docker exec -u rabbitmq "${RABBITMQ_CONTAINER_NAME}" rabbitmq-diagnostics -q check_port_connectivity > /dev/null 2>&1; then
      break
    fi
    sleep 1s
  done
  docker ps -a
}

stop_rabbitmq() {
  docker stop "${RABBITMQ_CONTAINER_NAME}"
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
