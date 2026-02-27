FROM golang:1.22-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY dataset ./dataset
# gen/ not needed: dataset has no proto imports

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /dataset-gen ./dataset

FROM gcr.io/distroless/static-debian12

COPY --from=builder /dataset-gen /dataset-gen

VOLUME ["/data"]
ENTRYPOINT ["/dataset-gen"]
