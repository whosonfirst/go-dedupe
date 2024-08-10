# go-whosonfirst-feature

A Go package for working with Who's On First (WOF) GeoJSON records.

## Important

This is exploratory work to develop a standard Go package for working with Who's On First (WOF) GeoJSON records that will eventually replace the `go-whosonfirst-geojson-v2` package.

If you are reading this that means the work is still ongoing and this package may still change at any time.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-whosonfirst-feature.svg)](https://pkg.go.dev/github.com/whosonfirst/go-whosonfirst-feature)

Documentation is incomplete at this time.

## History

In the beginning there was a `go-whosonfirst-geojson` package. And then, because it was created before the native Go module and versioning systems were finalized, there was a `go-whosonfirst-geojson-v2` package. Rather than resurrecting the `go-whosonfirst-geojson` package as `go-whosonfirst-geojson/v3` it was decided to start a new package namespace from scratch.

## First steps

The idea, so far, is to:

* Use the [paulmach/orb](https://github.com/paulmach/orb) package for working with geometries.
* Use the [tidwall/gjson](https://github.com/tidwall/gjson) package for querying properties.
* Writing custom code for deriving the meaning, or relevance, of properties.
* Probably defining a custom interface for WOF features. This is what the `go-whosonfirst-geojson-v2` package does. This package _should_ work with both WOF style GeoJSON records as well as "plain-vanilla" GeoJSON records, within the limits of that interface.

## Notes

* This package does not handle _formatting_ WOF records. That is handled by the [whosonfirst/go-whosonfirst-format](https://github.com/whosonfirst/go-whosonfirst-format) package.
* This package does not handle _validating_ WOF records. That is handled by the [whosonfirst/go-whosonfirst-validate](https://github.com/whosonfirst/go-whosonfirst-validate) package.
* This package does not handle _exporting_ (or writing) WOF records. That is handled by the [whosonfirst/go-whosonfirst-export](https://github.com/whosonfirst/go-whosonfirst-export) package.

## See also

* https://github.com/paulmach/orb
* https://github.com/tidwall/gjson
* https://github.com/whosonfirst/go-whosonfirst-format
* https://github.com/whosonfirst/go-whosonfirst-export
* https://github.com/whosonfirst/go-whosonfirst-validate