PREFIX	?= /usr/local
BIN	= goapl

all:
	go build -o $(BIN) ./cmd/goapl

install: all
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 755 $(BIN) $(DESTDIR)$(PREFIX)/bin/

clean:
	rm -f $(BIN)
	go clean

test:
	go vet ./...
	go test ./...

.PHONY: all install clean test
