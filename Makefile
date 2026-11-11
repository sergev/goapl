PREFIX	?= $(HOME)/.local
BIN	= goapl

all:
	go build -o $(BIN) ./cmd/goapl

install: all
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 755 $(BIN) $(DESTDIR)$(PREFIX)/bin/

clean:
	rm -rf $(BIN) build Manual.pdf
	go clean

test:
	go vet ./...
	go test ./...

# Compare with the Pascal prototype; needs fpc and its source directory.
compare: all
	@test -n "$(PROTO)" || { echo "usage: make compare PROTO=/path/to/pascal-apl"; exit 1; }
	python3 tools/compare.py $(PROTO) ./$(BIN) tools/compare_cases.txt

pdf: Manual.pdf

Manual.pdf: Manual.md tools/manual.typ
	pandoc Manual.md -o Manual.pdf --pdf-engine=typst -H tools/manual.typ

.PHONY: all install clean test compare pdf
