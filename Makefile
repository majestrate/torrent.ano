all:
	GOPATH=$(PWD) go build -o indextracker
	GOPATH=$(PWD) go build ./cmd/trackermanager
