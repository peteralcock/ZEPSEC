# ---- Build stage ----
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o zepsec-nmap-agent .

# ---- Runtime stage ----
FROM alpine:3.19

RUN apk add --no-cache \
    nmap \
    nmap-scripts \
    sudo \
    ca-certificates \
    tzdata

# Create non-root user with sudo nmap access
RUN addgroup -S zepsec && adduser -S zepsec -G zepsec && \
    echo "zepsec ALL=(root) NOPASSWD: /usr/bin/nmap" >> /etc/sudoers.d/zepsec

WORKDIR /app

COPY --from=builder /build/zepsec-nmap-agent .
RUN mkdir -p /tmp/zepsec-nmap && chown zepsec:zepsec /tmp/zepsec-nmap

USER zepsec

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget -qO- http://localhost:8000/health || exit 1

ENTRYPOINT ["./zepsec-nmap-agent"]
CMD ["-config", "/app/config.yml"]
