COMPILER=go
INDEXTRACKER=indextracker
TRACKERMANAGER=trackermanager

all:
	$(COMPILER) get -u -v .
	$(COMPILER) build -o $(INDEXTRACKER)
	$(COMPILER) build ./cmd/trackermanager

clean:
	go clean -a
	rm -f $(INDEXTRACKER) $(TRACKERMANAGER)
