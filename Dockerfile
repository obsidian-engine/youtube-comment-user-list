# マルチステージビルドでGoアプリケーションを最適化
FROM golang:1.24-alpine AS builder

# ====== Build Args ======
# ビルド対象のパッケージパス（例：./cmd/server, ./）
ARG APP_PATH=./cmd/server
ARG GOOS=linux
ARG GOARCH=amd64

# ====== 必要パッケージ ======
RUN apk add --no-cache git ca-certificates

# ====== 作業ディレクトリ ======
WORKDIR /app

# ====== 依存解決（キャッシュ最適化） ======
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# ====== ビルド（静的リンク + build cache） ======
# - cd は使わず、go build にパッケージパスを渡す
# - /root/.cache/go-build を buildx のキャッシュにマウント
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
      go build -trimpath -ldflags="-w -s -extldflags '-static'" -o /app/server ${APP_PATH}

# ====== 本番用の最小イメージ ======
FROM scratch

# CA証明書（HTTPS 通信で必要）
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# ビルドしたバイナリ
COPY --from=builder /app/server /server

EXPOSE 8080

# 実行
ENTRYPOINT ["/server"]