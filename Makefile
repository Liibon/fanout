.PHONY: generate build vet tidy clean

PROTOC_GEN_GO      := $(shell go env GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(shell go env GOPATH)/bin/protoc-gen-go-grpc

# Re-generate from proto (only needed after editing the .proto file).
generate: $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC)
	mkdir -p gen/hdsearchv1
	protoc \
		--go_out=. \
		--go_opt=module=github.com/liibon/fanout \
		--go-grpc_out=. \
		--go-grpc_opt=module=github.com/liibon/fanout \
		--proto_path=proto \
		proto/hdsearch/v1/hdsearch.proto

$(PROTOC_GEN_GO):
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.33.0

$(PROTOC_GEN_GO_GRPC):
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Build all pure-Go services (leaf requires Docker; see leaf/Dockerfile).
build:
	CGO_ENABLED=0 go build -o /tmp/root-svc   ./root
	CGO_ENABLED=0 go build -o /tmp/loadgen     ./loadgen
	CGO_ENABLED=0 go build -o /tmp/dataset-gen ./dataset

vet:
	go vet ./root/... ./loadgen/... ./dataset/...

tidy:
	go mod tidy

clean:
	rm -f /tmp/root-svc /tmp/loadgen /tmp/dataset-gen

demo: ## run the two-phase incast demo
	./demo-incast.sh

bench: ## run loadgen at 200 QPS for 10k requests
	docker compose run --rm loadgen -qps=200 -measure=10000

teardown: ## stop the stack and remove the dataset volume
	./scripts/teardown.sh
