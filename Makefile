MODULE   := github.com/inchix/intersight-mcp-server
BINARY   := intersight-mcp-server
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -s -w -X main.version=$(VERSION)
IMAGE    := ghcr.io/inchix/intersight-mcp-server

.PHONY: build run test lint clean docker docker-push

build:
	CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o $(BINARY) ./cmd/intersight-mcp-server

run: build
	./$(BINARY)

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -f $(BINARY)

docker:
	docker build -t $(IMAGE):$(VERSION) -t $(IMAGE):latest .

docker-push: docker
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest
