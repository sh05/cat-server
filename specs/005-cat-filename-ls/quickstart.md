# クイックスタート: ファイル内容取得エンドポイント

**機能**: `/cat/{filename}` エンドポイント
**目的**: 実装後の機能検証とデモンストレーション

## 前提条件

### 必要なツール
- Go 1.22+ (パスパラメータ機能に必要)
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

# テスト用テキストファイル作成
echo "Hello, World!" > ./files/hello.txt
echo -e "Line 1\nLine 2\nLine 3" > ./files/multiline.txt
echo '{"name": "config", "port": 8080}' > ./files/config.json
echo "# README\nThis is a test file." > ./files/README.md

# 隠しファイル作成
echo "SECRET_KEY=test123" > ./files/.env
echo "debug=true" > ./files/.config

# 空ファイル作成
touch ./files/empty.txt

# カスタムディレクトリ作成
mkdir -p ./custom-dir
echo "Custom content" > ./custom-dir/custom.txt
echo "Another file" > ./custom-dir/another.txt

# 特殊文字を含むファイル名
echo "Special content" > "./files/file with spaces.txt"
echo "日本語内容" > "./files/japanese-文字.txt"

# 大きなファイル作成（サイズ制限テスト用）
for i in {1..1000}; do
    echo "Line $i: This is a test line with some content to make the file larger." >> ./files/large.txt
done

# バイナリファイル作成（エラーテスト用）
echo -e '\x89PNG\r\n\x1a\n' > ./files/binary.bin
```

### 2. ファイル構成確認
```bash
# 作成されたファイル確認
ls -la ./files/
# 期待される出力:
# -rw-r--r-- hello.txt
# -rw-r--r-- multiline.txt
# -rw-r--r-- config.json
# -rw-r--r-- README.md
# -rw-r--r-- .env
# -rw-r--r-- .config
# -rw-r--r-- empty.txt
# -rw-r--r-- file with spaces.txt
# -rw-r--r-- japanese-文字.txt
# -rw-r--r-- large.txt
# -rw-r--r-- binary.bin

ls -la ./custom-dir/
# 期待される出力:
# -rw-r--r-- custom.txt
# -rw-r--r-- another.txt
```

## 機能デモンストレーション

### 1. サーバー起動（デフォルト設定）
```bash
# デフォルトディレクトリ (./files/) でサーバー起動
go run src/main.go

# 期待されるログ出力:
# {"level":"INFO","msg":"starting cat-server","directory":"./files/"}
# {"level":"INFO","msg":"server started successfully","addr":":8080"}
```

**別ターミナルで以下のテストを実行:**

### 2. 基本機能テスト
```bash
# 基本ファイル内容取得
curl -s http://localhost:8080/cat/hello.txt | jq .

# 期待されるレスポンス:
# {
#   "content": "Hello, World!",
#   "filename": "hello.txt",
#   "size": 13,
#   "directory": "./files/",
#   "generated_at": "2025-09-20T10:00:00Z"
# }
```

### 3. 複数行ファイルテスト
```bash
# 改行を含むファイルの内容取得
curl -s http://localhost:8080/cat/multiline.txt | jq .

# 期待されるレスポンス:
# {
#   "content": "Line 1\nLine 2\nLine 3",
#   "filename": "multiline.txt",
#   "size": 21,
#   "directory": "./files/",
#   "generated_at": "2025-09-20T10:00:00Z"
# }
```

### 4. JSONファイルテスト
```bash
# JSONファイルの内容取得
curl -s http://localhost:8080/cat/config.json | jq .

# 期待されるレスポンス:
# {
#   "content": "{\"name\": \"config\", \"port\": 8080}",
#   "filename": "config.json",
#   "size": 32,
#   "directory": "./files/",
#   "generated_at": "2025-09-20T10:00:00Z"
# }
```

### 5. 隠しファイルテスト
```bash
# 隠しファイルの内容取得
curl -s http://localhost:8080/cat/.env | jq .

# 期待されるレスポンス:
# {
#   "content": "SECRET_KEY=test123",
#   "filename": ".env",
#   "size": 18,
#   "directory": "./files/",
#   "generated_at": "2025-09-20T10:00:00Z"
# }
```

### 6. 空ファイルテスト
```bash
# 空ファイルの内容取得
curl -s http://localhost:8080/cat/empty.txt | jq .

# 期待されるレスポンス:
# {
#   "content": "",
#   "filename": "empty.txt",
#   "size": 0,
#   "directory": "./files/",
#   "generated_at": "2025-09-20T10:00:00Z"
# }
```

### 7. カスタムディレクトリテスト
```bash
# サーバー停止 (Ctrl+C)

# カスタムディレクトリでサーバー起動
go run src/main.go -dir ./custom-dir

# 新しいターミナルでテスト
curl -s http://localhost:8080/cat/custom.txt | jq .

# 期待されるレスポンス:
# {
#   "content": "Custom content",
#   "filename": "custom.txt",
#   "size": 14,
#   "directory": "./custom-dir/",
#   "generated_at": "2025-09-20T10:00:00Z"
# }
```

## エラーケーステスト

### 1. ファイル不存在エラー (404)
```bash
# 存在しないファイルの取得
curl -s http://localhost:8080/cat/nonexistent.txt | jq .

# 期待されるレスポンス:
# {
#   "error": "file not found",
#   "filename": "nonexistent.txt",
#   "path": "./files/nonexistent.txt",
#   "timestamp": "2025-09-20T10:00:00Z",
#   "status_code": 404
# }
```

### 2. パストラバーサル攻撃防止 (400)
```bash
# パストラバーサル攻撃の試行
curl -s http://localhost:8080/cat/../etc/passwd | jq .

# 期待されるレスポンス:
# {
#   "error": "invalid filename",
#   "filename": "../etc/passwd",
#   "timestamp": "2025-09-20T10:00:00Z",
#   "status_code": 400
# }
```

### 3. バイナリファイルエラー (415)
```bash
# バイナリファイルの取得試行
curl -s http://localhost:8080/cat/binary.bin | jq .

# 期待されるレスポンス:
# {
#   "error": "binary file not supported",
#   "filename": "binary.bin",
#   "path": "./files/binary.bin",
#   "timestamp": "2025-09-20T10:00:00Z",
#   "status_code": 415
# }
```

### 4. 大きなファイルエラー (413)
```bash
# 非常に大きなファイル作成 (10MB超)
dd if=/dev/zero of=./files/huge.txt bs=1M count=11

# 大きなファイルの取得試行
curl -s http://localhost:8080/cat/huge.txt | jq .

# 期待されるレスポンス:
# {
#   "error": "file too large",
#   "filename": "huge.txt",
#   "path": "./files/huge.txt",
#   "timestamp": "2025-09-20T10:00:00Z",
#   "status_code": 413
# }
```

### 5. HTTP メソッドテスト (405)
```bash
# POST メソッド (許可されない)
curl -X POST http://localhost:8080/cat/hello.txt

# 期待されるレスポンス: 405 Method Not Allowed

# PUT メソッド (許可されない)
curl -X PUT http://localhost:8080/cat/hello.txt

# 期待されるレスポンス: 405 Method Not Allowed
```

## パフォーマンステスト

### 1. レスポンス時間測定
```bash
# 小さなファイルのレスポンス時間測定
time curl -s http://localhost:8080/cat/hello.txt > /dev/null

# 期待される結果: < 50ms

# 中程度のファイルのレスポンス時間測定
time curl -s http://localhost:8080/cat/large.txt > /dev/null

# 期待される結果: < 200ms
```

### 2. 特殊文字ファイル名テスト
```bash
# スペースを含むファイル名
curl -s "http://localhost:8080/cat/file%20with%20spaces.txt" | jq .

# 日本語を含むファイル名
curl -s "http://localhost:8080/cat/japanese-%E6%96%87%E5%AD%97.txt" | jq .
```

## 統合テストシナリオ

### シナリオ1: 受け入れ条件1検証
**前提条件**: デフォルトディレクトリ (./files/) にファイルが存在
**実行**: `/cat/example.txt` エンドポイントにGETリクエスト
**期待結果**: ファイル内容がJSONで返される

```bash
# 実行
curl -s http://localhost:8080/cat/hello.txt

# 検証ポイント:
# 1. HTTP 200 OK
# 2. Content-Type: application/json
# 3. content フィールドにファイル内容
# 4. size と content のバイト数が一致
# 5. filename が正確
```

### シナリオ2: 受け入れ条件2検証
**前提条件**: `-dir /custom/path` でサーバー起動
**実行**: `/cat/config.json` エンドポイントにGETリクエスト
**期待結果**: カスタムパスのファイル内容が返される

```bash
# 実行
go run src/main.go -dir ./custom-dir
curl -s http://localhost:8080/cat/custom.txt

# 検証ポイント:
# 1. directory フィールドが "./custom-dir/"
# 2. カスタムディレクトリのファイル内容
# 3. デフォルトディレクトリのファイルにアクセス不可
```

### シナリオ3: セキュリティ検証
**前提条件**: パストラバーサル攻撃の試行
**実行**: `/cat/../../../etc/passwd` エンドポイントにGETリクエスト
**期待結果**: 400エラーが返される

```bash
# 実行
curl -s http://localhost:8080/cat/../secret.txt

# 検証ポイント:
# 1. HTTP 400 Bad Request
# 2. セキュリティエラーメッセージ
# 3. システムファイルにアクセス不可
```

## トラブルシューティング

### よくある問題

1. **サーバーが起動しない**
   ```bash
   # ポート使用確認
   lsof -i :8080
   # 使用中の場合は該当プロセス終了
   ```

2. **ファイルが見つからない**
   ```bash
   # ディレクトリ確認
   ls -la ./files/
   # ファイルパス確認
   ```

3. **権限エラー**
   ```bash
   # ファイル権限確認
   ls -l ./files/hello.txt
   # 読み取り権限があることを確認
   ```

4. **JSONパースエラー**
   ```bash
   # レスポンス確認
   curl -v http://localhost:8080/cat/hello.txt
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
rm -rf ./files ./custom-dir

# 大きなファイルの削除
rm -f ./files/huge.txt 2>/dev/null || true
```

---
**注意**: このクイックスタートは実装完了後に有効です。実装前は各ステップが失敗することが予想されます（TDDアプローチ）。