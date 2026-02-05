BINARY    := zepsec-nmap-agent
MODULE    := github.com/peteralcock/ZEPSEC/agents/go-nmap
VERSION   := 1.0.0
LDFLAGS   := -ldflags="-s -w -X main.version=$(VERSION)"
GOFLAGS   := CGO_ENABLED=0

.PHONY: build clean test docker run fmt vet

build:
	$(GOFLAGS) go build $(LDFLAGS) -o $(BINARY) .

build-linux:
	$(GOFLAGS) GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux-amd64 .

build-linux-arm:
	$(GOFLAGS) GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-linux-arm64 .

build-darwin:
	$(GOFLAGS) GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-darwin-amd64 .

build-windows:
	$(GOFLAGS) GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-windows-amd64.exe .

build-all: build-linux build-linux-arm build-darwin build-windows

docker:
	docker build -t $(BINARY):$(VERSION) .
	docker tag $(BINARY):$(VERSION) $(BINARY):latest

run:
	go run . -config config.yml

test:
	go test -v -race ./...

fmt:
	gofmt -s -w .

vet:
	go vet ./...

clean:
	rm -f $(BINARY) $(BINARY)-*
	go clean
