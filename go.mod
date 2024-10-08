module github.com/whosonfirst/go-dedupe

go 1.22.5

// This fixes a problem where io.Readers are not closed which results in filehandle exhaustion
replace github.com/philippgille/chromem-go v0.6.0 => github.com/philippgille/chromem-go v0.0.0-20240602150210-34ba8797e203

require (
	github.com/aaronland/go-jsonl v0.0.21
	github.com/aaronland/go-roster v1.0.0
	github.com/aaronland/gocloud-blob v0.3.0
	github.com/asg017/sqlite-vec-go-bindings v0.1.1
	github.com/blevesearch/bleve/v2 v2.4.2
	github.com/bwmarrin/snowflake v0.3.0
	github.com/json-iterator/go v1.1.12
	github.com/marcboeker/go-duckdb v1.7.1
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/mmcloughlin/geohash v0.10.0
	github.com/neuml/txtai.go v1.0.0
	github.com/ollama/ollama v0.3.6
	github.com/opensearch-project/opensearch-go/v2 v2.3.0
	github.com/paulmach/orb v0.11.1
	github.com/philippgille/chromem-go v0.6.0
	github.com/sfomuseum/go-csvdict v1.0.0
	github.com/sfomuseum/go-edtf v1.2.1
	github.com/sfomuseum/go-flags v0.10.0
	github.com/sfomuseum/go-timings v1.4.0
	github.com/tidwall/gjson v1.17.3
	github.com/whosonfirst/go-overture v0.0.0-20240812171335-eb48d41e6c6c
	github.com/whosonfirst/go-reader v1.0.2
	github.com/whosonfirst/go-whosonfirst-export/v2 v2.8.2
	github.com/whosonfirst/go-whosonfirst-feature v0.0.28
	github.com/whosonfirst/go-whosonfirst-iterate/v2 v2.5.0
	github.com/whosonfirst/go-whosonfirst-opensearch v0.0.18
	github.com/whosonfirst/go-whosonfirst-reader v1.0.2
	github.com/whosonfirst/go-whosonfirst-uri v1.3.0
	github.com/whosonfirst/go-whosonfirst-writer/v3 v3.1.4
	github.com/whosonfirst/go-writer/v3 v3.1.1
	gocloud.dev v0.39.0
)

require (
	github.com/RoaringBitmap/roaring v1.9.3 // indirect
	github.com/aaronland/go-artisanal-integers v0.9.1 // indirect
	github.com/aaronland/go-aws-auth v1.6.3 // indirect
	github.com/aaronland/go-brooklynintegers-api v1.2.7 // indirect
	github.com/aaronland/go-json-query v0.1.5 // indirect
	github.com/aaronland/go-pool/v2 v2.0.0 // indirect
	github.com/aaronland/go-string v1.0.0 // indirect
	github.com/aaronland/go-uid v0.4.0 // indirect
	github.com/aaronland/go-uid-artisanal v0.0.4 // indirect
	github.com/aaronland/go-uid-proxy v0.2.0 // indirect
	github.com/aaronland/go-uid-whosonfirst v0.0.5 // indirect
	github.com/apache/arrow/go/v17 v17.0.0 // indirect
	github.com/aws/aws-sdk-go-v2 v1.30.3 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.27.27 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.27 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.11 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.15 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.15 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/cognitoidentity v1.25.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/iam v1.34.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssm v1.52.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.22.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.26.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.30.3 // indirect
	github.com/aws/smithy-go v1.20.3 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/bits-and-blooms/bitset v1.12.0 // indirect
	github.com/blevesearch/bleve_index_api v1.1.10 // indirect
	github.com/blevesearch/geo v0.1.20 // indirect
	github.com/blevesearch/go-faiss v1.0.20 // indirect
	github.com/blevesearch/go-porterstemmer v1.0.3 // indirect
	github.com/blevesearch/gtreap v0.1.1 // indirect
	github.com/blevesearch/mmap-go v1.0.4 // indirect
	github.com/blevesearch/scorch_segment_api/v2 v2.2.15 // indirect
	github.com/blevesearch/segment v0.9.1 // indirect
	github.com/blevesearch/snowballstem v0.9.0 // indirect
	github.com/blevesearch/upsidedown_store_api v1.0.2 // indirect
	github.com/blevesearch/vellum v1.0.10 // indirect
	github.com/blevesearch/zapx/v11 v11.3.10 // indirect
	github.com/blevesearch/zapx/v12 v12.3.10 // indirect
	github.com/blevesearch/zapx/v13 v13.3.10 // indirect
	github.com/blevesearch/zapx/v14 v14.3.10 // indirect
	github.com/blevesearch/zapx/v15 v15.3.13 // indirect
	github.com/blevesearch/zapx/v16 v16.1.5 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/g8rswimmer/error-chain v1.0.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-resty/resty/v2 v2.3.0 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/flatbuffers v24.3.25+incompatible // indirect
	github.com/googleapis/gax-go/v2 v2.13.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/natefinch/atomic v1.0.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/sfomuseum/iso8601duration v1.1.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/whosonfirst/go-ioutil v1.0.2 // indirect
	github.com/whosonfirst/go-whosonfirst-crawl v0.2.2 // indirect
	github.com/whosonfirst/go-whosonfirst-flags v0.5.2 // indirect
	github.com/whosonfirst/go-whosonfirst-format v0.4.1 // indirect
	github.com/whosonfirst/go-whosonfirst-id v1.2.4 // indirect
	github.com/whosonfirst/go-whosonfirst-sources v0.1.0 // indirect
	github.com/whosonfirst/walk v0.0.2 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.etcd.io/bbolt v1.3.7 // indirect
	go.mongodb.org/mongo-driver v1.11.4 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.uber.org/ratelimit v0.3.0 // indirect
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225 // indirect
	golang.org/x/mod v0.18.0 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	golang.org/x/tools v0.22.0 // indirect
	golang.org/x/xerrors v0.0.0-20240716161551-93cc26a95ae9 // indirect
	google.golang.org/api v0.191.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240812133136-8ffd90a71988 // indirect
	google.golang.org/grpc v1.65.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)
