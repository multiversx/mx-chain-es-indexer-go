# Help section
help:
	@echo "Available targets:"
	@echo "  config        - Configure the local testnet"
	@echo "  start         - Start the local testnet"
	@echo "  stop          - Stop the local testnet"
	@echo "  clean         - Clean up the files of the local testnet"
	@echo "  cluster       - Start an Elasticsearch cluster on port 9200"
	@echo "  cluster-stop  - Stop the Elasticsearch cluster"
	@echo "  monitor-start - Start the Prometheus server (port 9090) and the Grafana server (port 3000)"
	@echo "  monitor-Stop  - Stop the Prometheus server (port 9090) and the Grafana server (port 3000)"

config:
	python3 config.py
start:
	python3 start.py
stop:
	python3 stop.py
clean:
	python3 clean.py
	cd .. && /bin/bash script.sh delete
cluster:
	cd .. && /bin/bash script.sh start
cluster-stop:
	cd .. && /bin/bash script.sh stop
monitor-start:
	cd .. && /bin/bash script.sh start_prometheus_and_grafana
monitor-stop:
	cd .. && /bin/bash script.sh stop_prometheus_and_grafana

