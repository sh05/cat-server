# クイックスタートガイド: ヘルスエンドポイント実装

**対象**: 開発者・運用チーム
**前提条件**: Go 1.21+ インストール済み
**推定時間**: 15-20分

## 概要

このガイドでは、cat-serverのヘルスチェックエンドポイント `/health` を実装し、テストするための手順を説明します。

## 1. プロジェクト初期化

### 1.1 プロジェクト構造作成

```bash
# プロジェクトルートで実行
mkdir -p src/server src/handlers tests/unit tests/integration
touch src/main.go src/handlers/health.go src/server/server.go
```

### 1.2 Go モジュール初期化

```bash
# go.mod 作成
go mod init github.com/sh05/cat-server

# 最低限の依存関係（標準ライブラリのみ）
echo 'module github.com/sh05/cat-server

go 1.21

require (
    // 標準ライブラリのみ使用
)' > go.mod
```

## 2. 基本実装

### 2.1 ヘルスハンドラー実装

`src/handlers/health.go` を作成：

```go
package handlers

import (
    "encoding/json"
    "log/slog"
    "net/http"
    "time"
)

// HealthResponse represents the health check response
type HealthResponse struct {
    Status    string    `json:"status"`
    Timestamp time.Time `json:"timestamp"`
}

// HealthHandler handles GET /health requests
func HealthHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()

    // Log the request
    slog.Info("health check requested",
        "remote_addr", r.RemoteAddr,
        "user_agent", r.UserAgent())

    // Create response
    response := HealthResponse{
        Status:    "ok",
        Timestamp: time.Now().UTC(),
    }

    // Set headers
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    // Encode and send response
    if err := json.NewEncoder(w).Encode(response); err != nil {
        slog.Error("failed to encode health response", "error", err)
        return
    }

    // Log completion
    duration := time.Since(start)
    slog.Info("health check completed",
        "duration", duration,
        "status", "ok")
}
```

### 2.2 サーバー実装

`src/server/server.go` を作成：

```go
package server

import (
    "context"
    "fmt"
    "log/slog"
    "net/http"
    "time"

    "github.com/sh05/cat-server/src/handlers"
)

// Server represents the HTTP server
type Server struct {
    httpServer *http.Server
    addr       string
}

// New creates a new server instance
func New(addr string) *Server {
    mux := http.NewServeMux()

    // Register health endpoint
    mux.HandleFunc("GET /health", handlers.HealthHandler)

    httpServer := &http.Server{
        Addr:         addr,
        Handler:      mux,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    return &Server{
        httpServer: httpServer,
        addr:       addr,
    }
}

// Start starts the HTTP server
func (s *Server) Start() error {
    slog.Info("starting server", "addr", s.addr)
    return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
    slog.Info("shutting down server")
    return s.httpServer.Shutdown(ctx)
}
```

### 2.3 メインアプリケーション

`src/main.go` を作成：

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "time"

    "github.com/sh05/cat-server/src/server"
)

func main() {
    // Setup structured logging
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    slog.SetDefault(logger)

    // Create server
    srv := server.New(":8080")

    // Setup graceful shutdown
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()

    // Start server in goroutine
    go func() {
        if err := srv.Start(); err != nil && err != http.ErrServerClosed {
            slog.Error("server failed to start", "error", err)
            os.Exit(1)
        }
    }()

    slog.Info("server started successfully", "addr", ":8080")

    // Wait for interrupt signal
    <-ctx.Done()

    // Shutdown server with timeout
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := srv.Shutdown(shutdownCtx); err != nil {
        slog.Error("server shutdown failed", "error", err)
        os.Exit(1)
    }

    slog.Info("server shutdown completed")
}
```

## 3. テスト実装

### 3.1 ユニットテスト

`tests/unit/health_test.go` を作成：

```go
package unit

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/sh05/cat-server/src/handlers"
)

func TestHealthHandler(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    w := httptest.NewRecorder()

    start := time.Now()
    handlers.HealthHandler(w, req)
    duration := time.Since(start)

    // Check status code
    if w.Code != http.StatusOK {
        t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
    }

    // Check content type
    expectedContentType := "application/json"
    if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
        t.Errorf("expected Content-Type %s, got %s", expectedContentType, contentType)
    }

    // Check response format
    var response handlers.HealthResponse
    if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }

    // Validate response fields
    if response.Status != "ok" {
        t.Errorf("expected status 'ok', got '%s'", response.Status)
    }

    if response.Timestamp.IsZero() {
        t.Error("expected non-zero timestamp")
    }

    // Check response time (should be fast)
    maxDuration := 10 * time.Millisecond
    if duration > maxDuration {
        t.Errorf("response took too long: %v > %v", duration, maxDuration)
    }
}
```

### 3.2 統合テスト

`tests/integration/health_integration_test.go` を作成：

```go
package integration

import (
    "encoding/json"
    "net/http"
    "testing"
    "time"

    "github.com/sh05/cat-server/src/handlers"
    "github.com/sh05/cat-server/src/server"
)

func TestHealthEndpointIntegration(t *testing.T) {
    // Start test server
    srv := server.New(":0") // Use random port

    go func() {
        if err := srv.Start(); err != nil && err != http.ErrServerClosed {
            t.Errorf("server failed to start: %v", err)
        }
    }()

    defer srv.Shutdown(context.Background())

    // Wait for server to start
    time.Sleep(100 * time.Millisecond)

    // Make request
    resp, err := http.Get("http://localhost:8080/health")
    if err != nil {
        t.Fatalf("failed to make request: %v", err)
    }
    defer resp.Body.Close()

    // Check status
    if resp.StatusCode != http.StatusOK {
        t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
    }

    // Check response
    var healthResp handlers.HealthResponse
    if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }

    if healthResp.Status != "ok" {
        t.Errorf("expected status 'ok', got '%s'", healthResp.Status)
    }
}
```

## 4. 実行とテスト

### 4.1 サーバー起動

```bash
# アプリケーション実行
go run src/main.go

# 出力例:
# {"time":"2025-09-20T10:15:30Z","level":"INFO","msg":"server started successfully","addr":":8080"}
```

### 4.2 ヘルスチェックテスト

別のターミナルで：

```bash
# cURLでテスト
curl -v http://localhost:8080/health

# 期待される出力:
# HTTP/1.1 200 OK
# Content-Type: application/json
# {"status":"ok","timestamp":"2025-09-20T10:15:30Z"}

# HTTPieでテスト (インストール済みの場合)
http GET localhost:8080/health

# Wgetでテスト
wget -qO- http://localhost:8080/health
```

### 4.3 テスト実行

```bash
# ユニットテスト実行
go test ./tests/unit/... -v

# 統合テスト実行
go test ./tests/integration/... -v

# 全テスト実行
go test ./... -v

# カバレッジ付きテスト
go test ./... -cover
```

### 4.4 品質ゲート実行

```bash
# コードフォーマット
go fmt ./...

# 静的解析
go vet ./...

# ビルド確認
go build -o bin/cat-server src/main.go

# 全品質ゲート実行
go vet ./... && go fmt ./... && go test ./... && go build -o bin/cat-server src/main.go
```

## 5. 負荷テスト

### 5.1 基本負荷テスト

```bash
# Apache Benchで100リクエスト、同時実行10
ab -n 100 -c 10 http://localhost:8080/health

# wrkで30秒間、2スレッド、10接続
wrk -t2 -c10 -d30s http://localhost:8080/health
```

### 5.2 期待される結果

- レスポンス時間: <10ms (95パーセンタイル)
- スループット: >1000 req/s
- エラー率: 0%
- メモリ使用量: <5MB

## 6. 監視設定例

### 6.1 cron監視スクリプト

```bash
#!/bin/bash
# /usr/local/bin/health-check.sh

ENDPOINT="http://localhost:8080/health"
TIMEOUT=5

if curl -f -s -m $TIMEOUT "$ENDPOINT" > /dev/null; then
    echo "$(date): Health check OK"
else
    echo "$(date): Health check FAILED" >&2
    exit 1
fi
```

### 6.2 systemd設定例

```ini
# /etc/systemd/system/cat-server.service
[Unit]
Description=cat-server REST API
After=network.target

[Service]
Type=simple
User=cat-server
ExecStart=/usr/local/bin/cat-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## 7. トラブルシューティング

### 7.1 よくある問題

**ポート使用中エラー**:
```bash
# ポート使用状況確認
lsof -i :8080
# または
netstat -tulpn | grep :8080
```

**Go バージョン互換性**:
```bash
# Go バージョン確認
go version
# Go 1.21+ が必要
```

**パーミッションエラー**:
```bash
# 1024未満のポート使用時はsudo必要
sudo ./cat-server  # ポート80/443使用時
```

### 7.2 ログ確認

```bash
# JSON形式ログの整形表示
./cat-server | jq .

# 特定レベルのログのみ表示
./cat-server | grep '"level":"ERROR"'
```

## 8. 次のステップ

1. **Docker化**: Dockerfile作成とコンテナ化
2. **CI/CD**: GitHub Actions等での自動テスト・デプロイ
3. **メトリクス**: Prometheus/Grafana監視設定
4. **セキュリティ**: HTTPS対応、セキュリティヘッダー追加

## 関連ドキュメント

- [実装計画](./plan.md)
- [データモデル](./data-model.md)
- [API仕様書](./contracts/health-api.yaml)
- [技術調査結果](./research.md)