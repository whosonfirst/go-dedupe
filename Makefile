GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

cli:
	go build -tags sqlite,sqlite_vec,duckdb,ollama,openclip -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/compare-locations cmd/compare-locations/main.go
	go build -tags sqlite,sqlite_vec,duckdb,ollama,openclip -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/index-locations cmd/index-locations/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/wof-assign-concordances cmd/wof-assign-concordances/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/wof-migrate-deprecated cmd/wof-migrate-deprecated/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/wof-process-duplicates cmd/wof-process-duplicates/main.go

# https://github.com/marcboeker/go-duckdb?tab=readme-ov-file#vendoring
modvendor:
	modvendor -copy="**/*.a **/*.h" -v
