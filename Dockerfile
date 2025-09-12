# ---- Build Stage ----
FROM golang:1.24-alpine AS builder

ARG APP_PATH=./cmd/server
ARG GOOS=linux
ARG GOARCH=amd64

RUN apk add --no-cache git ca-certificates
WORKDIR /app

# まずソース一式をコピー（go.work が存在すれば一緒に入る）
COPY . .

# 依存解決（go.work があれば自動的に考慮される）
RUN go mod download

# ビルド（静的リンク）
RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
    go build -trimpath -ldflags="-w -s -extldflags '-static'" \
    -o /app/server ${APP_PATH}

# ---- Runtime Stage ----
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/server /server

EXPOSE 8080
ENTRYPOINT ["/server"]