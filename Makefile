GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

cli:
	go build -mod $(GOMOD) -tags duckdb -ldflags="$(LDFLAGS)" -o bin/compare-locations cmd/compare-locations/main.go
	go build -mod $(GOMOD) -tags duckdb -ldflags="$(LDFLAGS)" -o bin/index-locations cmd/index-locations/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/wof-assign-concordances cmd/wof-assign-concordances/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/wof-migrate-deprecated cmd/wof-migrate-deprecated/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/wof-process-duplicates cmd/wof-process-duplicates/main.go

duckdb:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -tags duckdb  ./...
