PREFIX	?= $(HOME)/.local
BIN	= goapl

all:
	go build -o $(BIN) ./cmd/goapl

install: all
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 755 $(BIN) $(DESTDIR)$(PREFIX)/bin/

clean:
	rm -rf $(BIN) build
	go clean

test:
	go vet ./...
	go test ./...

# Compare with the Pascal prototype; needs fpc and its source directory.
compare: all
	@test -n "$(PROTO)" || { echo "usage: make compare PROTO=/path/to/pascal-apl"; exit 1; }
	python3 tools/compare.py $(PROTO) ./$(BIN) tools/compare_cases.txt

.PHONY: all install clean test compare
