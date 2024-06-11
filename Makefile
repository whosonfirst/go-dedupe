# https://github.com/aaronland/go-tools
URLESCAPE=$(shell which urlescape)

# Opensearch server

# This is for debugging. Do not change this at your own risk.
# (That means you should change this.)
OS_PSWD=KJHFGDFJGSJfsdkjfhsdoifruwo45978h52dcn

OS_MODEL=a8-aBJABf__qJekL_zJC
OS_BULK=false

OS_DSN="https://localhost:9200/dedupe?username=admin&password=$(OS_PSWD)&insecure=true&require-tls=true"
ENC_OS_DSN=$(shell $(URLESCAPE) $(OS_DSN))

OS_DATABASE_URI=opensearch://?dsn=$(ENC_OS_DSN)&model=$(OS_MODEL)&bulk-index=$(OS_BULK)

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
	| jq

# https://opensearch.org/docs/latest/ml-commons-plugin/pretrained-models/
# step 1: local-config-cluster-settings
# step 2: local-config-model-group
# step 3: local-config-register-model
# step 4: local-config-pipeline
# step 5: local-config-index

local-config-cluster-settings:
	cat database/opensearch/cluster-settings.json | \
		curl -k -s \
		-H 'Content-Type: application/json' \
		-X PUT \
		https://admin:$(OS_PSWD)@localhost:9200/_cluster/settings \
		-d @-

# To do: Extract model_group...

local-config-model-group:
	cat database/opensearch/local-model-group.json | \
		curl -k -s \
		-H 'Content-Type: application/json' \
		-X POST \
		https://admin:$(OS_PSWD)@localhost:9200/_plugins/_ml/model_groups/_register \
		-d @-

# To do: Write / insert model_group...

local-config-register-model:
	cat database/opensearch/register-model.json | \
		curl -k -s \
		-H 'Content-Type: application/json' \
		-X POST \
		https://admin:$(OS_PSWD)@localhost:9200/_plugins/_ml/models/_register \
		-d @-

# To do: Write / insert model_id...

local-config-deploy-model:
	cat database/opensearch/register-model.json | \
		curl -k -s \
		-H 'Content-Type: application/json' \
		-X POST \
		https://admin:$(OS_PSWD)@localhost:9200/_plugins/_ml/models/$(MODEL_ID)/_deploy \
		-d @-

# To do: Extract model_id...

local-task-status:
	curl -k -s \
	-H 'Content-Type: application/json' \
	-X GET \
	https://admin:$(OS_PSWD)@localhost:9200/_plugins/_ml/tasks/$(TASKID) \
	| jq

local-config-pipeline:
	cat database/opensearch/dedupe-ingest-pipeline.json | \
		curl -k -s \
		-H 'Content-Type: application/json' \
		-X PUT \
		https://admin:$(OS_PSWD)@localhost:9200/_ingest/pipeline/dedupe-ingest-pipeline \
		-d @-

local-config-index:
	cat database/opensearch/dedupe-index.json | \
	curl -k -s \
		-H 'Content-type:application/json' \
		-XPUT \
		https://admin:$(OS_PSWD)@localhost:9200/dedupe \
		-d @-

local-index-overture:
	go run cmd/index-overture-places/main.go \
		-database-uri "$(OS_DATABASE_URI)" \
		$(DATA)

local-query:
	go run cmd/query/main.go \
		-database-uri "$(OS_DATABASE_URI)" \
		-query "$(QUERY)"
