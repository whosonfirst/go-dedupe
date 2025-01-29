GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

cli:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/emit cmd/emit/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/reverse-geocode cmd/reverse-geocode/main.go
