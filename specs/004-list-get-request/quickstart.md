# クイックスタート: ファイル一覧取得エンドポイント

**機能**: `/ls` エンドポイント
**目的**: 実装後の機能検証とデモンストレーション

## 前提条件

### 必要なツール
- Go 1.21+ (最新安定版推奨)
- curl (HTTPリクエスト送信用)
- jq (JSON整形用、オプション)

### プロジェクト準備
```bash
# リポジトリルートディレクトリにいることを確認
pwd
# 出力例: /Users/username/ghq/github.com/sh05/cat-server

# 品質ゲート確認 (実装前は失敗することがある)
go vet ./...
go fmt ./...
go test ./...
go build ./...
```

## デモ環境セットアップ

### 1. テスト用ディレクトリとファイル作成
```bash
# デフォルトディレクトリ作成
mkdir -p ./files

# テスト用ファイル作成
echo "Hello World" > ./files/README.md
echo "test content" > ./files/test.txt
echo "hidden content" > ./files/.hidden
echo "config data" > ./files/.gitignore

# カスタムディレクトリ作成
mkdir -p ./custom-dir
echo "custom file" > ./custom-dir/custom.txt
echo "secret" > ./custom-dir/.env

# 空ディレクトリ作成
mkdir -p ./empty-dir

# 権限テスト用ディレクトリ (Unix/Linux/macOS のみ)
if [[ "$OSTYPE" != "msys" && "$OSTYPE" != "win32" ]]; then
    mkdir -p ./restricted
    echo "restricted file" > ./restricted/secret.txt
    chmod 000 ./restricted  # 読み取り権限削除
fi
```

### 2. ファイル構成確認
```bash
# 作成されたファイル確認
ls -la ./files/
# 期待される出力:
# -rw-r--r-- README.md
# -rw-r--r-- test.txt
# -rw-r--r-- .hidden
# -rw-r--r-- .gitignore

ls -la ./custom-dir/
# 期待される出力:
# -rw-r--r-- custom.txt
# -rw-r--r-- .env

ls -la ./empty-dir/
# 期待される出力: (ファイルなし)
```

## 機能デモンストレーション

### 1. サーバー起動（デフォルト設定）
```bash
# デフォルトディレクトリ (./files/) でサーバー起動
go run src/main.go

# 期待されるログ出力:
# {"level":"INFO","msg":"server started successfully","addr":":8080"}
```

**別ターミナルで以下のテストを実行:**

### 2. 基本機能テスト
```bash
# 基本ファイル一覧取得
curl -s http://localhost:8080/ls | jq .

# 期待されるレスポンス:
# {
#   "files": ["README.md", "test.txt", ".hidden", ".gitignore"],
#   "directory": "./files/",
#   "count": 4,
#   "generated_at": "2025-09-20T10:00:00Z"
# }
```

### 3. 隠しファイル確認テスト
```bash
# 隠しファイルが含まれていることを確認
curl -s http://localhost:8080/ls | jq '.files[] | select(startswith("."))'

# 期待される出力:
# ".hidden"
# ".gitignore"
```

### 4. レスポンス構造テスト
```bash
# count フィールドと files 配列長の一致確認
curl -s http://localhost:8080/ls | jq '{count: .count, actual_length: (.files | length)}'

# 期待される出力:
# {
#   "count": 4,
#   "actual_length": 4
# }
```

### 5. サーバー停止とカスタムディレクトリテスト
```bash
# サーバー停止 (Ctrl+C)

# カスタムディレクトリでサーバー起動
go run src/main.go -dir ./custom-dir

# 新しいターミナルでテスト
curl -s http://localhost:8080/ls | jq .

# 期待されるレスポンス:
# {
#   "files": ["custom.txt", ".env"],
#   "directory": "./custom-dir/",
#   "count": 2,
#   "generated_at": "2025-09-20T10:00:00Z"
# }
```

### 6. 空ディレクトリテスト
```bash
# サーバー停止後、空ディレクトリでテスト
go run src/main.go -dir ./empty-dir

# テスト実行
curl -s http://localhost:8080/ls | jq .

# 期待されるレスポンス:
# {
#   "files": [],
#   "directory": "./empty-dir/",
#   "count": 0,
#   "generated_at": "2025-09-20T10:00:00Z"
# }
```

## エラーケーステスト

### 1. 存在しないディレクトリ
```bash
# サーバー停止後、存在しないディレクトリで起動試行
go run src/main.go -dir ./nonexistent

# 期待される動作: サーバー起動時エラーまたは起動後400エラーレスポンス
```

### 2. 権限エラーテスト (Unix/Linux/macOS のみ)
```bash
# 読み取り権限のないディレクトリでテスト
if [[ "$OSTYPE" != "msys" && "$OSTYPE" != "win32" ]]; then
    go run src/main.go -dir ./restricted

    # テスト実行
    curl -s http://localhost:8080/ls | jq .

    # 期待されるレスポンス (403 エラー):
    # {
    #   "error": "permission denied",
    #   "path": "./restricted/",
    #   "timestamp": "2025-09-20T10:00:00Z",
    #   "status_code": 403
    # }
fi
```

### 3. HTTP メソッドテスト
```bash
# POST メソッド (許可されない)
curl -X POST http://localhost:8080/ls

# 期待されるレスポンス: 405 Method Not Allowed

# PUT メソッド (許可されない)
curl -X PUT http://localhost:8080/ls

# 期待されるレスポンス: 405 Method Not Allowed
```

## パフォーマンステスト

### 1. レスポンス時間測定
```bash
# レスポンス時間測定
time curl -s http://localhost:8080/ls > /dev/null

# 期待される結果: < 100ms (1000ファイル以下のディレクトリで)
```

### 2. 大量ファイルテスト (オプション)
```bash
# 大量ファイル作成 (テスト用)
mkdir -p ./large-dir
for i in {1..1000}; do
    echo "file $i" > "./large-dir/file$i.txt"
done

# 大量ファイルディレクトリでテスト
go run src/main.go -dir ./large-dir

# パフォーマンス測定
time curl -s http://localhost:8080/ls | jq '.count'

# 期待される結果:
# - count: 1000
# - レスポンス時間: < 100ms
```

## 統合テストシナリオ

### シナリオ1: 受け入れ条件1検証
**前提条件**: デフォルトディレクトリ (./files/) にファイルが存在
**実行**: `/ls` エンドポイントにGETリクエスト
**期待結果**: 全ファイル名（隠しファイル含む）がJSONで返される

```bash
# 実行
curl -s http://localhost:8080/ls

# 検証ポイント:
# 1. HTTP 200 OK
# 2. Content-Type: application/json
# 3. files 配列に全ファイル含有
# 4. 隠しファイル (.hidden, .gitignore) も含有
# 5. count と files 配列長が一致
```

### シナリオ2: 受け入れ条件2検証
**前提条件**: `-dir /custom/path` でサーバー起動
**実行**: `/ls` エンドポイントにGETリクエスト
**期待結果**: カスタムパスのファイル一覧が返される

```bash
# 実行
go run src/main.go -dir ./custom-dir
curl -s http://localhost:8080/ls

# 検証ポイント:
# 1. directory フィールドが "./custom-dir/"
# 2. カスタムディレクトリのファイルのみ含有
# 3. デフォルトディレクトリのファイルは含まれない
```

## トラブルシューティング

### よくある問題

1. **サーバーが起動しない**
   ```bash
   # ポート使用確認
   lsof -i :8080
   # 使用中の場合は該当プロセス終了
   ```

2. **権限エラー**
   ```bash
   # ディレクトリ権限確認
   ls -ld ./files/
   # 読み取り権限があることを確認
   ```

3. **JSONパースエラー**
   ```bash
   # レスポンス確認
   curl -v http://localhost:8080/ls
   # Content-Type ヘッダー確認
   ```

### 品質ゲート確認
```bash
# 実装完了後、必ず実行
go vet ./...      # 静的解析
go fmt ./...      # コード整形
go test ./...     # 全テスト実行
go build ./...    # コンパイル確認
```

## クリーンアップ

```bash
# テスト用ファイル削除
rm -rf ./files ./custom-dir ./empty-dir ./large-dir

# 権限制限解除 (Unix/Linux/macOS)
if [[ "$OSTYPE" != "msys" && "$OSTYPE" != "win32" ]]; then
    chmod 755 ./restricted 2>/dev/null || true
    rm -rf ./restricted
fi
```

---
**注意**: このクイックスタートは実装完了後に有効です。実装前は各ステップが失敗することが予想されます（TDDアプローチ）。