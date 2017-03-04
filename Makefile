all:
	GOPATH=$(PWD) go build -o indextracker -v tracker
