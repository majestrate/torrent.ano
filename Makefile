
COMPILER=go
INDEXTRACKER=indextracker
TRACKERMANAGER=trackermanager

all: $(INDEXTRACKER) $(TRACKERMANAGER)

$(INDEXTRACKER):
	$(COMPILER) build -o $(INDEXTRACKER)
$(TRACKERMANAGER):
	$(COMPILER) build ./cmd/trackermanager

clean:
	go clean -a
	rm -f $(INDEXTRACKER) $(TRACKERMANAGER)
