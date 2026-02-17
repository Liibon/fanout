FROM golang:1.22-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY gen ./gen
COPY root ./root

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /root-svc ./root

# ── runtime ──────────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /root-svc /root-svc

EXPOSE 50051 9102
ENTRYPOINT ["/root-svc"]
