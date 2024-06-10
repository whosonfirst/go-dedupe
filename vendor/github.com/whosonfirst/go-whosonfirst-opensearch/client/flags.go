package client

import (
	"flag"
)

var os_username string
var os_password string
var os_aws_uri string
var os_endpoint string
var os_insecure bool

func AppendClientFlags(fs *flag.FlagSet) {

	fs.StringVar(&os_username, "opensearch-username", "", "...")
	fs.StringVar(&os_aws_uri, "opensearch-aws-credentials-uri", "", "...")
	fs.StringVar(&os_password, "opensearch-password", "", "...")
	fs.BoolVar(&os_insecure, "opensearch-insecure", false, "...")
	fs.StringVar(&os_endpoint, "opensearch-endpoint", "https://localhost:9200", "...")
}
