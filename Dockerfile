# マルチステージビルドでGoアプリケーションを最適化
FROM golang:1.24-alpine AS builder

# 必要なパッケージをインストール
RUN apk add --no-cache git ca-certificates

# 作業ディレクトリを設定
WORKDIR /app

# Go モジュールファイルをコピー
COPY go.mod go.sum ./

# 依存関係をダウンロード
RUN go mod download

# ソースコードをコピー
COPY . .

# バイナリをビルド（静的リンク & buildxキャッシュ対応）
RUN --mount=type=cache,target=/go/pkg/mod \
    cd cmd/server && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
      go build -ldflags="-w -s -extldflags '-static'" -a -installsuffix cgo -o /app/server .

# 本番用の最小イメージ
FROM scratch

# CA証明書をコピー（HTTPSリクエスト用）
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# ビルドしたバイナリをコピー
COPY --from=builder /app/server /server

# ポート8080を公開
EXPOSE 8080

# サーバーを起動
CMD ["/server"]