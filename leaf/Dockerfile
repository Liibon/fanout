FROM ubuntu:22.04 AS faiss-builder

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential cmake libopenblas-dev git ca-certificates && \
    rm -rf /var/lib/apt/lists/*

ARG FAISS_TAG=v1.7.3
RUN git clone --depth 1 --branch ${FAISS_TAG} \
        https://github.com/facebookresearch/faiss.git /faiss

WORKDIR /faiss
RUN cmake -B build \
        -DFAISS_ENABLE_GPU=OFF \
        -DFAISS_ENABLE_PYTHON=OFF \
        -DBUILD_TESTING=OFF \
        -DFAISS_ENABLE_C_API=ON \
        -DCMAKE_BUILD_TYPE=Release \
        -DBUILD_SHARED_LIBS=ON \
        .

FROM golang:1.22-bookworm AS go-builder
RUN apt-get update && apt-get install -y --no-install-recommends libopenblas-dev && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY gen ./gen
COPY leaf ./leaf
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /leaf-svc ./leaf

FROM ubuntu:22.04
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends libopenblas-base ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=go-builder /leaf-svc /leaf-svc
EXPOSE 50051
ENTRYPOINT ["/leaf-svc"]
