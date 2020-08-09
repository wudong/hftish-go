MAIN = hftish-go

COMPILE_COMMAND = go build -o bin/${MAIN} main/main.go

# Set source dir and scan source dir for all go files
SRC_DIR = .
SOURCES = $(shell find $(SRC_DIR) -type f -name '*.go')

BINARIES = $(wildcard bin/*)

build: $(SOURCES)
	$(COMPILE_COMMAND)

build-all-binaries: $(SOURCES) clean
	GOOS=darwin    GOARCH=amd64    $(COMPILE_COMMAND) && mv ./bin/${MAIN} ./bin/${MAIN}-darwin-amd64
	GOOS=linux     GOARCH=amd64    $(COMPILE_COMMAND) && mv ./bin/${MAIN} ./bin/${MAIN}-linux-amd64
	GOOS=windows   GOARCH=amd64    $(COMPILE_COMMAND) && mv ./bin/${MAIN} ./bin/${MAIN}-windows-amd64

compress-all-binaries: build-all-binaries
	for f in $(BINARIES); do      \
        tar czf $$f.tar.gz $$f;    \
    done
	@rm $(BINARIES)

test: $(SOURCES)
	@go vet .
	@test -z $(shell gofmt -s -l . | tee /dev/stderr) || (echo "[ERROR] Fix formatting issues with 'gofmt'" && exit 1)
	@go test

.PHONY: clean
clean:
	rm -Rf bin;
