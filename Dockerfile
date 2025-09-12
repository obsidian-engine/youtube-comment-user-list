# ---- Build Stage ----
FROM golang:1.24-alpine AS builder

# 任意: ビルド対象パス（mainパッケージ）を指定。既定は ./ （リポジトリ直下）
ARG APP_PATH=./
ARG GOOS=linux
ARG GOARCH=amd64

# 必要パッケージ
RUN apk add --no-cache git ca-certificates

# 作業ディレクトリ
WORKDIR /app

# go.mod / go.sum（go.work があれば一緒に）
COPY go.mod go.sum ./
# go.work を使っている場合だけ存在するのでワイルドカードで拾う
COPY go.work* . 2>/dev/null || true

# 依存解決
RUN go mod download

# 残りのソース
COPY . .

# ビルド（静的リンク）
RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
    go build -trimpath -ldflags="-w -s -extldflags '-static'" \
    -o /app/server ${APP_PATH}

# ---- Runtime Stage ----
FROM scratch

# CA証明書（HTTPS用）
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 実行バイナリ
COPY --from=builder /app/server /server

EXPOSE 8080
ENTRYPOINT ["/server"]