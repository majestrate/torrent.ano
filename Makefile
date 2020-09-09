PREFIX=GOPATH=$(PWD)
COMPILER=go
INDEXTRACKER=indextracker
TRACKERMANAGER=trackermanager
SUBMODULE=.submodule_built

all: $(SUBMODULE) $(INDEXTRACKER) $(TRACKERMANAGER)

$(INDEXTRACKER):
	$(PREFIX) $(COMPILER) build -o $(INDEXTRACKER)
$(TRACKERMANAGER):
	$(PREFIX) $(COMPILER) build ./cmd/trackermanager

$(SUBMODULE):
	git submodule update --init --recursive

clean:
	$(PREFIX) go clean -a
	rm -f $(INDEXTRACKER) $(TRACKERMANAGER)

distclean: clean
	rm -f $(SUBMODULE)
