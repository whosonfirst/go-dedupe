build:
	CGO_LDFLAGS="-Lwork/vss -Wl,-undefined,dynamic_lookup" go build ./...

vss:
	CGO_LDFLAGS="-Lwork/vss -Wl,-undefined,dynamic_lookup" go run -tags="json1" cmd/vss/main.go
