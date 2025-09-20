# クイックスタート: Docker Image作成機能

**フェーズ1出力** | **日付**: 2025-09-20

## 概要
このガイドは、Docker Image作成機能の実装完了後に、開発者がcat-serverのDockerイメージをビルド・実行するための手順を提供します。

## 前提条件
- Docker Engine 20.10+
- cat-serverのソースコード
- Linux/AMD64 環境 (推奨)

## 基本的な使用手順

### 1. Dockerイメージのビルド
```bash
# リポジトリルートディレクトリから実行
docker build -t cat-server:latest .

# ビルド結果の確認
docker images cat-server
```

**期待される結果**:
- ビルド時間: < 60秒 (初回)
- イメージサイズ: < 50MB
- エラーなしでの完了

### 2. コンテナの実行
```bash
# デフォルト設定での実行
docker run -d --name cat-server -p 8080:8080 cat-server:latest

# コンテナの状態確認
docker ps
```

**期待される結果**:
- 起動時間: < 2秒
- ステータス: Running
- ポート8080でのアクセス可能

### 3. アプリケーションの動作確認
```bash
# ヘルスチェック
curl http://localhost:8080/health

# ファイルリスト取得
curl http://localhost:8080/ls

# 特定ファイルの内容取得 (例: go.mod)
curl http://localhost:8080/cat/go.mod
```

**期待される結果**:
- /health: JSON形式のレスポンス
- /ls: ファイルリストのJSON
- /cat/{filename}: ファイル内容のJSON

### 4. コンテナの停止とクリーンアップ
```bash
# コンテナの停止
docker stop cat-server

# コンテナの削除
docker rm cat-server

# イメージの削除 (必要に応じて)
docker rmi cat-server:latest
```

## 高度な使用例

### カスタムディレクトリでの実行
```bash
# ホストディレクトリをマウントして実行
docker run -d --name cat-server \
  -p 8080:8080 \
  -v /path/to/host/directory:/app/files \
  cat-server:latest -dir /app/files

# マウントしたディレクトリのファイル確認
curl http://localhost:8080/ls
```

### 環境変数での設定
```bash
# 環境変数でポート変更
docker run -d --name cat-server \
  -p 9090:9090 \
  -e PORT=9090 \
  cat-server:latest

# 確認
curl http://localhost:9090/health
```

### Docker Composeでの実行
```yaml
# docker-compose.yml (実装後に作成予定)
version: '3.8'
services:
  cat-server:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./files:/app/files
    environment:
      - PORT=8080
      - GIN_MODE=release
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
```

```bash
# Docker Composeでの起動
docker-compose up -d

# ログの確認
docker-compose logs cat-server

# 停止
docker-compose down
```

## 検証とテスト

### パフォーマンステスト
```bash
# イメージサイズの確認
docker images cat-server --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"

# 起動時間の測定
time docker run --rm cat-server:latest --help

# メモリ使用量の確認
docker stats cat-server --no-stream
```

### セキュリティテスト
```bash
# ユーザー権限の確認
docker run --rm cat-server:latest whoami
# 期待結果: app (非root)

# 脆弱性スキャン (Docker Desktopがある場合)
docker scout cves cat-server:latest

# プロセス確認
docker exec cat-server ps aux
```

### ヘルスチェックテスト
```bash
# ヘルスチェック状態の確認
docker inspect cat-server --format='{{.State.Health.Status}}'

# ヘルスチェック履歴
docker inspect cat-server --format='{{range .State.Health.Log}}{{.Output}}{{end}}'
```

## トラブルシューティング

### よくある問題と解決方法

#### 1. ビルドエラー
```bash
# ビルドログの詳細確認
docker build --progress=plain -t cat-server:latest .

# キャッシュを無効にして再ビルド
docker build --no-cache -t cat-server:latest .
```

#### 2. 起動エラー
```bash
# コンテナログの確認
docker logs cat-server

# 対話的デバッグ
docker run -it --rm cat-server:latest /bin/sh
```

#### 3. ポート競合
```bash
# 使用可能ポートの確認
netstat -tulpn | grep :8080

# 別ポートでの実行
docker run -d --name cat-server -p 8081:8080 cat-server:latest
```

#### 4. ディスク容量不足
```bash
# Dockerの使用量確認
docker system df

# 不要なリソースのクリーンアップ
docker system prune -f
```

## パフォーマンス最適化

### ビルド最適化
```bash
# マルチステージビルドの各段階確認
docker build --target builder -t cat-server:builder .
docker images cat-server:builder

# レイヤーサイズの分析
docker history cat-server:latest
```

### 実行時最適化
```bash
# リソース制限付きでの実行
docker run -d --name cat-server \
  --memory="64m" \
  --cpus="0.1" \
  -p 8080:8080 \
  cat-server:latest

# リソース使用量の監視
docker stats cat-server
```

## 本番環境での運用

### ログ管理
```bash
# ログローテーション設定
docker run -d --name cat-server \
  --log-driver json-file \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  -p 8080:8080 \
  cat-server:latest
```

### 監視設定
```bash
# ヘルスチェック結果の定期確認
while true; do
  echo "$(date): $(docker inspect cat-server --format='{{.State.Health.Status}}')"
  sleep 30
done
```

## 次のステップ
実装完了後は、以下の拡張を検討してください：
- CI/CD パイプラインへの統合
- Multi-platform ビルド (ARM64サポート)
- Docker Registryへの自動プッシュ
- Kubernetes/OpenShift対応
- セキュリティスキャンの自動化