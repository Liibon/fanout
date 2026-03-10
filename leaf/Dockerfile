FROM ubuntu:22.04 AS faiss-builder

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential cmake libopenblas-dev git ca-certificates && \
    rm -rf /var/lib/apt/lists/*

ARG FAISS_TAG=v1.7.3
RUN git clone --depth 1 --branch ${FAISS_TAG} \
        https://github.com/facebookresearch/faiss.git /faiss

FROM golang:1.22-bookworm AS go-builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY gen ./gen
COPY leaf ./leaf
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /leaf-svc ./leaf

FROM ubuntu:22.04
COPY --from=go-builder /leaf-svc /leaf-svc
EXPOSE 50051
ENTRYPOINT ["/leaf-svc"]
