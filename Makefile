build:
	CGO_LDFLAGS="-Lwork -Wl,-undefined,dynamic_lookup" go build ./...
