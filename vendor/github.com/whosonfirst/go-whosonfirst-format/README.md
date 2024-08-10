# go-whosonfirst-format

Standardised GeoJSON formatting for Whos On First files.

Usable as both a library and a binary.

## Library usage

```golang
func main() {
  inputBytes, err := ioutil.ReadFile(inputPath)
  if err != nil {
    panic(err)
  }

  var feature format.Feature

  json.Unmarshal(inputBytes, &feature)
  if err != nil {
    panic(err)
  }

  outputBytes, err := format.FormatFeature(&feature)
  if err != nil {
    panic(err)
  }

  fmt.Printf("%s", outputBytes)
}
```

## Binary usage

```shell
make build
cat input.geojson | ./build/wof-format > output.geojson
```