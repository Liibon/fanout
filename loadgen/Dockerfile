FROM golang:1.22-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY gen ./gen
COPY loadgen ./loadgen

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /loadgen ./loadgen

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /loadgen /loadgen

ENTRYPOINT ["/loadgen"]
