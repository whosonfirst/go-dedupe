# https://github.com/aaronland/go-tools
URLESCAPE=$(shell which urlescape)

# Opensearch server

# This is for debugging. Do not change this at your own risk.
# (That means you should change this.)
OS_PSWD=KJHFGDFJGSJfsdkjfhsdoifruwo45978h52dcn

OS_DSN="https://localhost:9200/dedupe?username=admin&password=$(OS_PSWD)&insecure=true&require-tls=true"
ENC_OS_DSN=$(shell $(URLESCAPE) $(OS_DSN))

OS_DATABASE_URI=opensearch://?dsn=$(ENC_OS_DSN)

# https://opensearch.org/docs/latest/install-and-configure/install-opensearch/docker/
#
# And then:
# curl -v -k https://admin:$(OS_PSWD)@localhost:9200/

local-server:
	docker run \
		-it \
		-p 9200:9200 \
		-p 9600:9600 \
		-e "discovery.type=single-node" \
		-e "OPENSEARCH_INITIAL_ADMIN_PASSWORD=$(OS_PSWD)" \
		-v opensearch-data1:/usr/local/data/opensearch \
		opensearchproject/opensearch:latest

local-config:
	@make local-config-pipeline
	@make local-config-index

local-config-pipeline:
	cat database/opensearch/dedupe-ingest-pipeline.json | \
		curl -k \
		-H 'Content-Type: application/json' \
		-X PUT \
		https://admin:$(OS_PSWD)@localhost:9200/_ingest/pipeline/dedupe-ingest-pipeline \
		-d @-

local-config-index:
	cat database/opensearch/dedupe-index.json | \
	curl -k \
		-H 'Content-type:application/json' \
		-XPUT \
		https://admin:$(OS_PSWD)@localhost:9200/dedupe \
		-d @-

local-index-overture:
	go run cmd/index-overture-places/main.go \
		-database-uri $(OS_DATABASE_URI) \
		$(DATA)
