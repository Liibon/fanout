FROM golang:1.22-bookworm AS go-builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY gen ./gen
COPY leaf ./leaf

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /leaf-svc ./leaf

FROM ubuntu:22.04
COPY --from=go-builder /leaf-svc /leaf-svc
EXPOSE 50051
ENTRYPOINT ["/leaf-svc"]
