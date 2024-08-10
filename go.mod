module github.com/whosonfirst/go-dedupe

go 1.22.5

// This fixes a problem where io.Readers are not closed which results in filehandle exhaustion
replace github.com/philippgille/chromem-go v0.6.0 => github.com/philippgille/chromem-go v0.0.0-20240602150210-34ba8797e203

require (
	github.com/aaronland/go-jsonl v0.0.21
	github.com/aaronland/go-roster v1.0.0
	github.com/aaronland/gocloud-blob v0.0.17
	github.com/asg017/sqlite-vec-go-bindings v0.1.0
	github.com/blevesearch/bleve/v2 v2.4.1
	github.com/bwmarrin/snowflake v0.3.0
	github.com/json-iterator/go v1.1.12
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/mmcloughlin/geohash v0.10.0
	github.com/neuml/txtai.go v1.0.0
	github.com/ollama/ollama v0.2.1
	github.com/opensearch-project/opensearch-go/v2 v2.3.0
	github.com/paulmach/orb v0.11.1
	github.com/philippgille/chromem-go v0.6.0
	github.com/sfomuseum/go-csvdict v1.0.0
	github.com/sfomuseum/go-flags v0.10.0
	github.com/sfomuseum/go-timings v1.3.0
	github.com/tidwall/gjson v1.17.3
	github.com/whosonfirst/go-overture v0.0.0-20240608205656-2f9a190af0d5
	github.com/whosonfirst/go-reader v1.0.2
	github.com/whosonfirst/go-whosonfirst-feature v0.0.28
	github.com/whosonfirst/go-whosonfirst-iterate/v2 v2.4.1
	github.com/whosonfirst/go-whosonfirst-opensearch v0.0.18
	github.com/whosonfirst/go-whosonfirst-reader v1.0.2
	github.com/whosonfirst/go-writer-jsonl/v3 v3.0.1
	github.com/whosonfirst/go-writer/v3 v3.1.1
	gocloud.dev v0.38.0
)

require (
	github.com/RoaringBitmap/roaring v1.9.3 // indirect
	github.com/aaronland/go-aws-auth v1.6.1 // indirect
	github.com/aaronland/go-json-query v0.1.5 // indirect
	github.com/aws/aws-sdk-go-v2 v1.26.1 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.27.11 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.11 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/cognitoidentity v1.23.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/iam v1.31.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssm v1.50.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.23.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.6 // indirect
	github.com/aws/smithy-go v1.20.2 // indirect
	github.com/bits-and-blooms/bitset v1.12.0 // indirect
	github.com/blevesearch/bleve_index_api v1.1.9 // indirect
	github.com/blevesearch/geo v0.1.20 // indirect
	github.com/blevesearch/go-faiss v1.0.19 // indirect
	github.com/blevesearch/go-porterstemmer v1.0.3 // indirect
	github.com/blevesearch/gtreap v0.1.1 // indirect
	github.com/blevesearch/mmap-go v1.0.4 // indirect
	github.com/blevesearch/scorch_segment_api/v2 v2.2.14 // indirect
	github.com/blevesearch/segment v0.9.1 // indirect
	github.com/blevesearch/snowballstem v0.9.0 // indirect
	github.com/blevesearch/upsidedown_store_api v1.0.2 // indirect
	github.com/blevesearch/vellum v1.0.10 // indirect
	github.com/blevesearch/zapx/v11 v11.3.10 // indirect
	github.com/blevesearch/zapx/v12 v12.3.10 // indirect
	github.com/blevesearch/zapx/v13 v13.3.10 // indirect
	github.com/blevesearch/zapx/v14 v14.3.10 // indirect
	github.com/blevesearch/zapx/v15 v15.3.13 // indirect
	github.com/blevesearch/zapx/v16 v16.1.4 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/g8rswimmer/error-chain v1.0.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-resty/resty/v2 v2.3.0 // indirect
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/googleapis/gax-go/v2 v2.12.3 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/natefinch/atomic v1.0.1 // indirect
	github.com/sfomuseum/go-edtf v1.2.1 // indirect
	github.com/sfomuseum/iso8601duration v1.1.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/whosonfirst/go-ioutil v1.0.2 // indirect
	github.com/whosonfirst/go-whosonfirst-crawl v0.2.2 // indirect
	github.com/whosonfirst/go-whosonfirst-flags v0.5.1 // indirect
	github.com/whosonfirst/go-whosonfirst-sources v0.1.0 // indirect
	github.com/whosonfirst/go-whosonfirst-uri v1.3.0 // indirect
	github.com/whosonfirst/walk v0.0.2 // indirect
	go.etcd.io/bbolt v1.3.7 // indirect
	go.mongodb.org/mongo-driver v1.11.4 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/crypto v0.25.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	google.golang.org/api v0.176.1 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240415180920-8c6c420018be // indirect
	google.golang.org/grpc v1.63.2 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)
