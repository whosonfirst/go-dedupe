# go-timings

Go package implementing interface and methods for background timing monitors

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/sfomuseum/go-timings.svg)](https://pkg.go.dev/github.com/sfomuseum/go-timings)

## Exmaple

```
import (
       "context"
       "github.com/sfomuseum/go-timings"
       "os"       
)

func main() {

	ctx := context.Background()
	
	monitor, _ := timings.NewMonitor(ctx, "counter://PT60S)

	monitor.Start(ctx, os.Stdout)
	defer monitor.Stop(ctx)

	monitor.Signal(ctx)	// increments by 1
}
```
