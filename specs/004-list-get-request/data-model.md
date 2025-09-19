# データモデル: ファイル一覧取得エンドポイント

**作成日**: 2025-09-20
**機能**: `/list` エンドポイント

## エンティティ定義

### 1. DirectoryPath (ディレクトリパス)
**説明**: ファイル一覧取得対象のディレクトリパス

**属性**:
- `path` (string): ディレクトリの絶対パスまたは相対パス
- `isValid` (bool): パスの有効性フラグ
- `exists` (bool): ディレクトリ存在フラグ
- `readable` (bool): 読み取り権限フラグ

**検証ルール**:
- `path` は空文字列禁止
- `path` にnullバイト (`\x00`) 含有禁止
- `path` はディレクトリでなければならない（ファイル不可）
- パストラバーサル攻撃防止: `../` 連続使用制限
- 最大パス長: 4096文字以下（Unix/Linux 標準）

**状態遷移**:
```
初期状態 → 検証中 → 有効/無効
有効 → アクセス確認 → 読み取り可能/不可
```

**Go実装イメージ**:
```go
type DirectoryPath struct {
    Path     string
    IsValid  bool
    Exists   bool
    Readable bool
}

func (dp *DirectoryPath) Validate() error {
    // validation logic
}
```

### 2. FileList (ファイル一覧)
**説明**: 指定ディレクトリ内のファイル名コレクション

**属性**:
- `files` ([]string): ファイル名の配列
- `count` (int): ファイル数
- `includeHidden` (bool): 隠しファイル含有フラグ
- `directoryPath` (string): 対象ディレクトリパス
- `generatedAt` (time.Time): 一覧生成時刻

**検証ルール**:
- `files` 配列は nil 禁止（空配列は許可）
- 各ファイル名は有効なファイル名である必要がある
- 隠しファイル（ドット開始）は `includeHidden=true` 時のみ含む
- 最大ファイル数: 10,000個（メモリ保護）
- ファイル名最大長: 255文字（Unix/Linux 標準）

**関係**:
- DirectoryPath に 1対1 で関連
- 1つの DirectoryPath から 1つの FileList が生成される

**Go実装イメージ**:
```go
type FileList struct {
    Files         []string  `json:"files"`
    Count         int       `json:"count"`
    IncludeHidden bool      `json:"include_hidden"`
    DirectoryPath string    `json:"directory"`
    GeneratedAt   time.Time `json:"generated_at"`
}

func (fl *FileList) AddFile(filename string) {
    // add file with validation
}
```

### 3. APIResponse (API レスポンス)
**説明**: クライアントに返されるHTTP レスポンスデータ構造

#### 3.1 成功レスポンス (SuccessResponse)
**HTTP Status**: 200 OK

**属性**:
- `files` ([]string): ファイル名配列
- `directory` (string): 対象ディレクトリパス
- `count` (int): ファイル数
- `generated_at` (string): ISO 8601形式の生成時刻

**JSON形式**:
```json
{
  "files": ["file1.txt", ".hidden", "README.md"],
  "directory": "./files/",
  "count": 3,
  "generated_at": "2025-09-20T10:00:00Z"
}
```

**検証ルール**:
- すべてのフィールドは必須
- `files` は空配列でも有効
- `count` は `files` 配列長と一致する必要がある
- `generated_at` は RFC 3339 形式

#### 3.2 エラーレスポンス (ErrorResponse)
**HTTP Status**: 400/403/500

**属性**:
- `error` (string): エラーメッセージ
- `path` (string): 問題となったパス（該当する場合）
- `timestamp` (string): エラー発生時刻
- `status_code` (int): HTTP ステータスコード

**JSON形式**:
```json
{
  "error": "directory not found",
  "path": "/invalid/path",
  "timestamp": "2025-09-20T10:00:00Z",
  "status_code": 400
}
```

**エラー種別**:
- `400 Bad Request`: 無効なディレクトリパス
- `403 Forbidden`: ディレクトリ読み取り権限なし
- `500 Internal Server Error`: システムエラー

**Go実装イメージ**:
```go
type SuccessResponse struct {
    Files       []string  `json:"files"`
    Directory   string    `json:"directory"`
    Count       int       `json:"count"`
    GeneratedAt time.Time `json:"generated_at"`
}

type ErrorResponse struct {
    Error      string    `json:"error"`
    Path       string    `json:"path,omitempty"`
    Timestamp  time.Time `json:"timestamp"`
    StatusCode int       `json:"status_code"`
}
```

## データフロー

```
コマンドライン引数 → DirectoryPath
       ↓
   パス検証・アクセス確認
       ↓
   os.ReadDir() → FileList
       ↓
   JSON形式変換 → APIResponse
       ↓
   HTTP レスポンス送信
```

## パフォーマンス考慮事項

### メモリ使用量
- **小規模** (100ファイル): ~10KB
- **中規模** (1,000ファイル): ~100KB
- **大規模** (10,000ファイル): ~1MB

### 処理時間
- **目標**: ディレクトリアクセス + JSON変換 < 50ms
- **制限**: 10,000ファイル超過時はエラーレスポンス

## セキュリティ制約

### 入力検証
- パストラバーサル攻撃防止
- nullバイト インジェクション防止
- パス長制限によるバッファオーバーフロー防止

### アクセス制御
- 指定ディレクトリ外のアクセス禁止
- システムディレクトリ アクセス制限
- 読み取り権限事前確認

---
**次工程**: contracts/ 生成、OpenAPI仕様定義