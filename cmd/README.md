# Command line tools

## wof-assign-concordances

```
> go run cmd/assign-wof-concordances/main.go -h
  -concordance-as-int
    	If true cast the concordance ID as an int64
  -concordance-namespace string
    	The namespace of the concordance being applied.
  -concordance-predicate string
    	The predicate of the concordance being applies. (default "id")
  -mark-is-current
    	If true the addition of a cocordance will mark this record as mz:is_current=1
  -reader-uri string
    	A valid whosonfirst/go-reader URI for reading WOF records from.
  -verbose
    	Enable verbose (debug) logging.
  -whosonfirst-label string
    	The "label" used to identify WOF records. Valid options are: source, target. (default "target")
  -writer-uri string
    	A valid whosonfirst/go-reader URI for writing WOF records from.
```

For example:

```
$> go run cmd/assign-wof-concordances/main.go \
	-reader-uri repo:///usr/local/data/whosonfirst-data-venue-us-ny \
	-writer-uri repo:///usr/local/data/whosonfirst-data-venue-us-ny \
	-concordance-namespace ovtr \
	-concordance-predicate id \
	/usr/local/data/ovtr-wof-ny.csv
```