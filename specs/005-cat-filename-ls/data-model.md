# データモデル: ファイル内容取得エンドポイント

**作成日**: 2025-09-20
**機能**: `/cat/{filename}` エンドポイント

## エンティティ定義

### 1. FileRequest (ファイルリクエスト)
**説明**: HTTPリクエストから抽出されるファイル取得要求

**属性**:
- `filename` (string): URLパスパラメータから取得されるファイル名
- `directory` (string): サーバー起動時の`-dir`フラグで指定されたディレクトリパス
- `fullPath` (string): ディレクトリとファイル名を結合した完全パス

**検証ルール**:
- `filename` は空文字列禁止
- `filename` にnullバイト (`\0`) 含有禁止
- `filename` にパストラバーサル文字列 (`../`, `..\\`) 含有禁止
- `fullPath` は指定ディレクトリ配下でなければならない
- 最大ファイル名長: 255文字以下（Unix/Linux 標準）

**Go実装イメージ**:
```go
type FileRequest struct {
    Filename  string
    Directory string
    FullPath  string
}

func (fr *FileRequest) Validate() error {
    // validation logic
}
```

### 2. FileContent (ファイル内容)
**説明**: 読み取り対象ファイルの内容とメタデータ

**属性**:
- `content` (string): ファイルの内容（UTF-8テキスト）
- `size` (int64): ファイルサイズ（バイト数）
- `isText` (bool): テキストファイル判定フラグ
- `encoding` (string): エンコーディング種別（UTF-8固定）

**検証ルール**:
- `content` はUTF-8有効文字列でなければならない
- `size` は10MB（10485760バイト）以下でなければならない
- `isText` がfalseの場合はエラーとする
- バイナリファイルの読み込みを拒否

**状態遷移**:
```
ファイル検出 → サイズチェック → テキスト判定 → 内容読み込み → 検証完了
```

**Go実装イメージ**:
```go
type FileContent struct {
    Content  string `json:"content"`
    Size     int64  `json:"size"`
    IsText   bool   `json:"-"`
    Encoding string `json:"encoding"`
}

func (fc *FileContent) LoadFromFile(path string) error {
    // file loading and validation logic
}
```

### 3. CatResponse (Cat API レスポンス)
**説明**: クライアントに返されるHTTP レスポンスデータ構造

#### 3.1 成功レスポンス (SuccessResponse)
**HTTP Status**: 200 OK

**属性**:
- `content` (string): ファイルの内容
- `filename` (string): 要求されたファイル名
- `size` (int64): ファイルサイズ（バイト数）
- `directory` (string): 対象ディレクトリパス
- `generated_at` (string): ISO 8601形式の生成時刻

**JSON形式**:
```json
{
  "content": "Hello, World!\nThis is a sample file.",
  "filename": "example.txt",
  "size": 34,
  "directory": "./files/",
  "generated_at": "2025-09-20T10:00:00Z"
}
```

**検証ルール**:
- すべてのフィールドは必須
- `content` は改行・特殊文字を含む可能性がある
- `size` は `content` のバイト数と一致する必要がある
- `generated_at` は RFC 3339 形式

#### 3.2 エラーレスポンス (ErrorResponse)
**HTTP Status**: 400/403/404/413/415/500

**属性**:
- `error` (string): エラーメッセージ
- `filename` (string): 問題となったファイル名
- `path` (string): 問題となったパス（該当する場合）
- `timestamp` (string): エラー発生時刻
- `status_code` (int): HTTP ステータスコード

**JSON形式**:
```json
{
  "error": "file not found",
  "filename": "nonexistent.txt",
  "path": "./files/nonexistent.txt",
  "timestamp": "2025-09-20T10:00:00Z",
  "status_code": 404
}
```

**エラー種別**:
- `400 Bad Request`: 無効なファイル名、パストラバーサル攻撃
- `403 Forbidden`: ファイル読み取り権限なし
- `404 Not Found`: ファイルが存在しない
- `413 Payload Too Large`: ファイルサイズ制限超過（>10MB）
- `415 Unsupported Media Type`: バイナリファイル
- `500 Internal Server Error`: システムエラー

**Go実装イメージ**:
```go
type CatSuccessResponse struct {
    Content     string    `json:"content"`
    Filename    string    `json:"filename"`
    Size        int64     `json:"size"`
    Directory   string    `json:"directory"`
    GeneratedAt time.Time `json:"generated_at"`
}

type CatErrorResponse struct {
    Error      string    `json:"error"`
    Filename   string    `json:"filename,omitempty"`
    Path       string    `json:"path,omitempty"`
    Timestamp  time.Time `json:"timestamp"`
    StatusCode int       `json:"status_code"`
}
```

## エンティティ関係

```
FileRequest (1) → validates → (1) FileContent
FileContent (1) → transforms → (1) CatSuccessResponse
ValidationError → transforms → (1) CatErrorResponse
```

**関係性の説明**:
1. `FileRequest` が検証され、有効な場合のみ `FileContent` の読み込みを実行
2. `FileContent` が正常に読み込まれた場合、`CatSuccessResponse` に変換
3. 任意の段階でエラーが発生した場合、`CatErrorResponse` に変換

## データフロー

```
HTTP Request → FileRequest → Validation → File Access
     ↓
File Content Loading → Text Validation → Size Check
     ↓
CatSuccessResponse → JSON Serialization → HTTP Response
```

**エラーフロー**:
```
Validation Error → CatErrorResponse → HTTP Error Response
File Not Found → CatErrorResponse → HTTP 404
Permission Error → CatErrorResponse → HTTP 403
Size Limit Exceeded → CatErrorResponse → HTTP 413
Binary File → CatErrorResponse → HTTP 415
System Error → CatErrorResponse → HTTP 500
```

## パフォーマンス考慮事項

### メモリ使用量
- **小ファイル** (1KB): ~2KB（メタデータ含む）
- **中ファイル** (100KB): ~200KB（メタデータ含む）
- **大ファイル** (10MB): ~20MB（JSON変換時の一時的な倍増）

### 処理時間目標
- **1KB以下**: <10ms
- **100KB以下**: <50ms
- **10MB以下**: <200ms
- **制限超過**: 即座に413エラー

## セキュリティ制約

### パストラバーサル防止
- ファイル名に `../` または `..\\` を含む場合は400エラー
- `filepath.Clean()` による正規化後の再チェック
- 最終パスが指定ディレクトリ配下かの確認

### ファイルアクセス制御
- 読み取り権限の事前確認
- 指定ディレクトリ外のアクセス完全禁止
- シンボリックリンクによる迂回防止

### コンテンツ制御
- バイナリファイルの読み込み拒否
- ファイルサイズ制限による DoS 攻撃防止
- UTF-8エンコーディング強制

---
**次工程**: OpenAPI仕様定義（contracts/），コントラクトテスト作成