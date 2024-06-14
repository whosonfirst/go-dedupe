# https://github.com/aaronland/go-tools
URLESCAPE=$(shell which urlescape)

# Opensearch server

# This is for debugging. Do not change this at your own risk.
# (That means you should change this.)
OS_PSWD=KJHFGDFJGSJfsdkjfhsdoifruwo45978h52dcn

OS_MODEL=9dgHD5ABSoo-6k3cWDqn

# If true this tends to trigger Java heap memory errors for OS run inside a Docker container
OS_BULK=false

OS_DSN="https://localhost:9200/dedupe?username=admin&password=$(OS_PSWD)&insecure=true&require-tls=true"
ENC_OS_DSN=$(shell $(URLESCAPE) $(OS_DSN))

OS_DATABASE_URI=opensearch://?dsn=$(ENC_OS_DSN)&model=$(OS_MODEL)&bulk-index=$(OS_BULK)

# https://opensearch.org/docs/latest/install-and-configure/install-opensearch/docker/
#
# And then:
# curl -v -k https://admin:$(OS_PSWD)@localhost:9200/

#		-it \

local-server:
	docker run \
		-p 9200:9200 \
		-p 9600:9600 \
		-e "discovery.type=single-node" \
		-e "OPENSEARCH_INITIAL_ADMIN_PASSWORD=$(OS_PSWD)" \
		-v opensearch-data1:/usr/local/data/opensearch \
		opensearchproject/opensearch:latest

# Quick and dirty targets for testing things

local-aliases:
	curl -k -s \
	-H 'Content-Type: application/json' \
	-X GET \
	https://admin:$(OS_PSWD)@localhost:9200/_aliases \
	| jq

local-settings:
	curl -k -s \
	-H 'Content-Type: application/json' \
	-X GET \
	https://admin:$(OS_PSWD)@localhost:9200/dedupe/_settings \
	| jq

local-mappings:
	curl -k -s \
	-H 'Content-Type: application/json' \
	-X GET \
	https://admin:$(OS_PSWD)@localhost:9200/dedupe/_mappings \
	| jq

local-search:
	curl -k -s \
	-H 'Content-Type: application/json' \
	-X GET \
	https://admin:$(OS_PSWD)@localhost:9200/dedupe/_search \
	| json_pp 

# https://opensearch.org/docs/latest/ml-commons-plugin/pretrained-models/
# https://opensearch.org/blog/improving-document-retrieval-with-sparse-semantic-encoders/

# step 1: local-config-cluster-settings
# step 3: local-config-register-model
# step 4: local-task-status TASK=<task-id> and record <model-id>
# step 5: write <model-id> to static/opensearch/dedupe-ingest-pipeline-*.json 
# step 6: local-config-pipeline
# step 7: local-config-index
# step 8: local-index-overture DATA=<path>

local-config-cluster-settings:
	cat static/opensearch/cluster-settings.json | \
		curl -k -s \
		-H 'Content-Type: application/json' \
		-X PUT \
		https://admin:$(OS_PSWD)@localhost:9200/_cluster/settings \
		-d @-


local-config-register-model:
	cat static/opensearch/register-model-text.json | \
		curl -k -s \
		-H 'Content-Type: application/json' \
		-X POST \
		'https://admin:$(OS_PSWD)@localhost:9200/_plugins/_ml/models/_register?deploy=true' \
		-d @-

# To do: Extract model_id...

local-task-status:
	curl -k -s \
	-H 'Content-Type: application/json' \
	-X GET \
	https://admin:$(OS_PSWD)@localhost:9200/_plugins/_ml/tasks/$(TASK) \
	| jq

# To do: Write / insert model_id..

local-config-pipeline:
	cat static/opensearch/dedupe-ingest-pipeline-text.json | \
		curl -k -s \
		-H 'Content-Type: application/json' \
		-X PUT \
		https://admin:$(OS_PSWD)@localhost:9200/_ingest/pipeline/dedupe-ingest-pipeline \
		-d @-

local-remove-index:
	curl -k -s \
		-H 'Content-type:application/json' \
		-XDELETE \
		https://admin:$(OS_PSWD)@localhost:9200/dedupe

local-config-index:
	cat static/opensearch/dedupe-index-text.json | \
	curl -k -s \
		-H 'Content-type:application/json' \
		-XPUT \
		https://admin:$(OS_PSWD)@localhost:9200/dedupe \
		-d @-

# -start-after 2518764 \

local-index-overture:
	go run cmd/index-overture-places/main.go \
		-database-uri "$(OS_DATABASE_URI)" \
		$(DATA)

local-query:
	go run cmd/query/main.go \
		-database-uri "$(OS_DATABASE_URI)" \
		-name "$(NAME)" \
		-address "$(ADDRESS)" $(METADATA)	
