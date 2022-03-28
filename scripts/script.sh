IMAGE_NAME=elastic-container
DEFAULT_ES_VERSION=7.16.2

start() {
  ES_VERSION=$1
  if [ -z "${ES_VERSION}" ]; then
    ES_VERSION=${DEFAULT_ES_VERSION}
  fi

  docker pull docker.elastic.co/elasticsearch/elasticsearch:${ES_VERSION}

  docker rm ${IMAGE_NAME} 2> /dev/null
  docker run -d --name "${IMAGE_NAME}" -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:${ES_VERSION}

  # Wait elastic cluster to start
  echo "Waiting Elasticsearch cluster to start..."
  sleep 30s
}

stop() {
  docker stop "${IMAGE_NAME}"
}

delete() {
  curl -XDELETE http://localhost:9200/_all

  curl -XDELETE http://localhost:9200/_template/*
}

"$@"
