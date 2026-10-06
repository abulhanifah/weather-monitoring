# ---------- builder ----------
FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /app/server ./cmd

# ---------- runtime ----------
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata wget \
  && adduser -D -u 10001 appuser
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD wget -qO- http://127.0.0.1:${APP_PORT:-8080}/health || exit 1
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/app/server"]
