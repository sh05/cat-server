# クイックスタート: Go標準ディレクトリ構造リファクタリング

**機能**: 007-go-cmd-pkg
**目的**: リファクタリング後の機能検証とデモンストレーション

## 前提条件

- Go 1.21+ がインストールされている
- Docker がインストールされている（オプション）
- cat-server リポジトリが手元にある

## 1. リファクタリング前後の構造比較

### リファクタリング前（現在の構造）
```
cat-server/
├── src/
│   ├── main.go
│   ├── server/
│   ├── handlers/
│   └── services/
├── tests/
└── go.mod
```

### リファクタリング後（目標構造）
```
cat-server/
├── cmd/
│   └── cat-server/
│       └── main.go
├── pkg/
│   ├── domain/
│   ├── application/
│   ├── infrastructure/
│   └── interfaces/
├── internal/
│   ├── config/
│   └── utils/
├── tests/
└── go.mod
```

## 2. 基本的な動作確認

### 2.1 品質ゲートの実行

リファクタリング前後で以下のコマンドが全て成功することを確認：

```bash
# 静的解析
go vet ./...

# コードフォーマット確認
go fmt ./...

# 全テスト実行
go test ./...

# ビルド確認
go build ./cmd/cat-server
```

**期待結果**: 全てのコマンドがエラーなしで完了

### 2.2 アプリケーションの起動

```bash
# リファクタリング後の起動
go run ./cmd/cat-server/main.go

# または、ビルドしてから実行
go build -o bin/cat-server ./cmd/cat-server
./bin/cat-server
```

**期待結果**:
```
{"level":"INFO","msg":"starting cat-server","directory":"./files/","time":"2025-09-20T..."}
{"level":"INFO","msg":"server started successfully","addr":":8080","time":"2025-09-20T..."}
```

## 3. API機能確認テスト

### 3.1 ヘルスチェックエンドポイント

```bash
# 基本ヘルスチェック
curl http://localhost:8080/health

# JSON形式での確認
curl -H "Accept: application/json" http://localhost:8080/health

# HTML形式での確認
curl -H "Accept: text/html" http://localhost:8080/health
```

**期待結果**:
```json
{
  "status": "healthy",
  "timestamp": "2025-09-20T...",
  "version": "1.0.0"
}
```

### 3.2 ファイルリストエンドポイント

```bash
# ディレクトリ一覧取得
curl http://localhost:8080/ls

# JSON形式での詳細確認
curl -s http://localhost:8080/ls | jq .
```

**期待結果**:
```json
{
  "files": [
    {
      "name": "go.mod",
      "size": 123,
      "modTime": "2025-09-20T...",
      "isDir": false
    }
  ],
  "directory": "./files/",
  "totalCount": 1
}
```

### 3.3 ファイル内容エンドポイント

```bash
# 特定ファイルの内容取得
curl http://localhost:8080/cat/go.mod

# JSON形式での確認
curl -s http://localhost:8080/cat/go.mod | jq .
```

**期待結果**:
```json
{
  "filename": "go.mod",
  "content": "module github.com/sh05/cat-server\n\ngo 1.21",
  "size": 45,
  "contentType": "text/plain"
}
```

## 4. エラーケースの確認

### 4.1 セキュリティテスト

```bash
# パストラバーサル攻撃の防止確認
curl http://localhost:8080/cat/../etc/passwd
# 期待結果: 400 Bad Request

# 存在しないファイルのアクセス
curl http://localhost:8080/cat/nonexistent.txt
# 期待結果: 404 Not Found
```

### 4.2 HTTPメソッドの制限

```bash
# POST メソッドでのアクセス（許可されていない）
curl -X POST http://localhost:8080/ls
# 期待結果: 405 Method Not Allowed

curl -X POST http://localhost:8080/cat/go.mod
# 期待結果: 405 Method Not Allowed
```

## 5. パフォーマンス確認

### 5.1 レスポンス時間測定

```bash
# ヘルスチェックの応答時間
curl -w "%{time_total}\n" -s -o /dev/null http://localhost:8080/health

# ファイルリストの応答時間
curl -w "%{time_total}\n" -s -o /dev/null http://localhost:8080/ls

# ファイル取得の応答時間
curl -w "%{time_total}\n" -s -o /dev/null http://localhost:8080/cat/go.mod
```

**期待結果**: 全ての応答時間が 200ms 以下

### 5.2 同時接続テスト

```bash
# 10並行リクエストでの負荷テスト
for i in {1..10}; do
  curl -s http://localhost:8080/health &
done
wait
```

**期待結果**: 全てのリクエストが正常に完了

## 6. Docker環境での確認

### 6.1 Dockerイメージのビルド

```bash
# イメージビルド
docker build -t cat-server:refactored .

# または、スクリプトを使用
./scripts/docker-build.sh refactored
```

**期待結果**: イメージが正常にビルドされ、サイズが50MB以下

### 6.2 コンテナでの動作確認

```bash
# コンテナ起動
docker run -d --name cat-server-test -p 8080:8080 cat-server:refactored

# 動作確認
curl http://localhost:8080/health

# 非ルートユーザー確認
docker exec cat-server-test whoami
# 期待結果: app

# ログ確認
docker logs cat-server-test

# クリーンアップ
docker stop cat-server-test
docker rm cat-server-test
```

## 7. コントラクトテストの実行

### 7.1 リファクタリング検証テスト

```bash
# リファクタリング専用のコントラクトテスト実行
go test ./specs/007-go-cmd-pkg/contracts/ -v

# 特定のテストケースの実行
go test ./specs/007-go-cmd-pkg/contracts/ -run TestRefactoringValidationContract -v
```

**期待結果**: 全てのテストが PASS

### 7.2 既存機能の回帰テスト

```bash
# 全ての既存テストが引き続き動作することを確認
go test ./tests/unit/... -v
go test ./tests/integration/... -v
go test ./tests/contract/... -v
```

**期待結果**: 既存テスト全体で失敗なし

## 8. トラブルシューティング

### 8.1 よくある問題

**問題**: `go build` でパッケージが見つからない
```bash
# 解決: モジュールの再初期化
go mod tidy
go mod download
```

**問題**: テストが失敗する
```bash
# 解決: 詳細なテスト出力で原因調査
go test -v -race ./...
```

**問題**: インポートパスエラー
```bash
# 解決: go.mod ファイルの確認とインポートパスの修正
cat go.mod
# module名とインポートパスの整合性を確認
```

### 8.2 リファクタリング検証チェックリスト

- [ ] 新しいディレクトリ構造（cmd/, pkg/, internal/）が存在する
- [ ] 旧 src/ ディレクトリが完全に削除されている
- [ ] 全ての品質ゲート（go vet, go fmt, go test, go build）が成功
- [ ] 全てのAPIエンドポイントが正常に動作
- [ ] Docker環境での動作確認完了
- [ ] パフォーマンス指標の維持確認
- [ ] セキュリティ機能の維持確認

## 9. 成功基準

リファクタリングが成功したと判断する基準：

1. **機能的互換性**: 全てのAPIが既存と同じ動作をする
2. **品質維持**: 全ての品質ゲートがパス
3. **構造準拠**: Go標準ディレクトリ構造に準拠
4. **パフォーマンス**: 応答時間とリソース使用量が同等以下
5. **テスト**: 既存テストに加えて新しいコントラクトテストが成功
6. **清掃完了**: 古いコード/ディレクトリが完全に削除されている

---

**最終更新**: 2025-09-20
**テスト実行時間**: 約5-10分