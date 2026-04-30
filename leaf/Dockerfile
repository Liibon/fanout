# ── Stage 1: build FAISS from source ─────────────────────────────────────────
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
        . && \
    cmake --build build --target faiss_c -j"$(nproc)" && \
    cmake --install build --prefix /faiss-install && \
    find build -name "libfaiss_c.so*" -exec cp -P {} /faiss-install/lib/ \;

# ── Stage 2: build Go binary ──────────────────────────────────────────────────
FROM golang:1.22-bookworm AS go-builder

RUN apt-get update && apt-get install -y --no-install-recommends \
        libopenblas-dev && \
    rm -rf /var/lib/apt/lists/*

COPY --from=faiss-builder /faiss-install /faiss-install

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY gen ./gen
COPY leaf ./leaf

RUN CGO_ENABLED=1 \
    CGO_CFLAGS="-I/faiss-install/include" \
    CGO_LDFLAGS="-L/faiss-install/lib -lfaiss_c -lopenblas -lstdc++ -lm" \
    go build -trimpath -ldflags="-s -w" -o /leaf-svc ./leaf

# ── Stage 3: runtime ──────────────────────────────────────────────────────────
FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
        libopenblas-base libgomp1 ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=faiss-builder /faiss-install/lib/libfaiss_c.so* /usr/local/lib/
COPY --from=faiss-builder /faiss-install/lib/libfaiss.so* /usr/local/lib/
RUN ldconfig

COPY --from=go-builder /leaf-svc /leaf-svc

EXPOSE 50051 9103
ENTRYPOINT ["/leaf-svc"]
