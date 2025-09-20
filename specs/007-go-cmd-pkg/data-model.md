# データモデル: Go標準ディレクトリ構造リファクタリング

**日付**: 2025-09-20
**プロジェクト**: cat-server リファクタリング (007-go-cmd-pkg)

## ドメインエンティティ

### 1. FileSystemEntry (ファイルシステムエントリ)

**説明**: ファイルシステム上のファイルまたはディレクトリを表現する中核エンティティ

**属性**:
- `name`: string - ファイル/ディレクトリ名
- `path`: string - 絶対パス
- `size`: int64 - ファイルサイズ（ディレクトリの場合は0）
- `modTime`: time.Time - 最終更新時刻
- `isDir`: bool - ディレクトリかどうかのフラグ
- `permissions`: os.FileMode - ファイル権限

**検証ルール**:
- `name` は空文字列不可
- `path` は有効な絶対パス形式必須
- `size` は非負数
- `permissions` は有効なファイル権限値

**ビジネスルール**:
- セキュリティ: パストラバーサル攻撃の防止
- 隠しファイル（.で始まる）のアクセス制御

### 2. DirectoryListing (ディレクトリリスト)

**説明**: ディレクトリ内のファイル一覧を表現する集約エンティティ

**属性**:
- `path`: string - 対象ディレクトリパス
- `entries`: []FileSystemEntry - ファイルエントリのコレクション
- `totalCount`: int - 総ファイル数
- `scannedAt`: time.Time - スキャン実行時刻

**検証ルール**:
- `path` は存在するディレクトリパス
- `entries` はnilでない
- `totalCount` はentriesの要素数と一致

**ビジネスルール**:
- アクセス権限の事前チェック
- 読み取り可能ディレクトリのみ処理

### 3. FileContent (ファイル内容)

**説明**: ファイルの内容とメタデータを表現するエンティティ

**属性**:
- `entry`: FileSystemEntry - ファイルエンティティの参照
- `content`: []byte - ファイルの生バイナリ内容
- `encoding`: string - 文字エンコーディング（推定）
- `readAt`: time.Time - 読み取り実行時刻

**検証ルール**:
- `entry` はファイル（非ディレクトリ）
- `content` のサイズ制限（メモリ保護）
- `encoding` は有効なエンコーディング名

**ビジネスルール**:
- 大容量ファイルの読み取り制限
- バイナリファイルの適切な処理

## 値オブジェクト

### 1. FilePath (ファイルパス)

**説明**: ファイルパスを表現する不変値オブジェクト

**属性**:
- `value`: string - 正規化されたパス文字列

**検証ルール**:
- パストラバーサル防止（../ 禁止）
- 絶対パス形式の強制
- 有効な文字セット

**操作**:
- `IsSecure()`: bool - セキュリティチェック
- `Join(relativePath)`: FilePath - パス結合
- `Base()`: string - ベース名の取得

### 2. FileSize (ファイルサイズ)

**説明**: ファイルサイズを表現する値オブジェクト

**属性**:
- `bytes`: int64 - バイト数

**検証ルール**:
- 非負数
- 最大サイズ制限

**操作**:
- `HumanReadable()`: string - 人間可読形式（1.2MB等）
- `IsLarge()`: bool - 大容量ファイル判定

## リポジトリインターフェース

### 1. FileSystemRepository

**説明**: ファイルシステムアクセスの抽象化インターフェース

**操作**:
```go
type FileSystemRepository interface {
    // ディレクトリ一覧の取得
    ListDirectory(path FilePath) (DirectoryListing, error)

    // ファイル内容の読み取り
    ReadFile(path FilePath) (FileContent, error)

    // ファイル存在確認
    Exists(path FilePath) bool

    // アクセス権限チェック
    IsReadable(path FilePath) bool

    // パスの検証
    ValidatePath(path string) error
}
```

### 2. HealthRepository

**説明**: ヘルスチェック機能の抽象化インターフェース

**操作**:
```go
type HealthRepository interface {
    // システムヘルス状態の取得
    GetHealthStatus() HealthStatus

    // 依存サービスのチェック
    CheckDependencies() []DependencyStatus
}
```

## アプリケーションサービス

### 1. DirectoryService

**説明**: ディレクトリ操作のユースケースを実装

**操作**:
- `GetDirectoryListing(path string)`: ディレクトリ一覧の安全な取得
- `ValidateDirectoryAccess(path string)`: アクセス権限の検証

### 2. FileService

**説明**: ファイル操作のユースケースを実装

**操作**:
- `GetFileContent(filename string)`: ファイル内容の安全な取得
- `ValidateFileAccess(path string)`: ファイルアクセスの検証

### 3. HealthService

**説明**: ヘルスチェックのユースケースを実装

**操作**:
- `GetSystemHealth()`: システム全体の健康状態
- `GetDetailedHealth()`: 詳細な診断情報

## 状態遷移

### ファイルアクセス状態
```
[リクエスト] → [パス検証] → [権限チェック] → [ファイル読み取り] → [レスポンス]
              ↓              ↓               ↓
            [エラー:        [エラー:         [エラー:
             無効パス]       権限なし]        読み取り失敗]
```

### ディレクトリスキャン状態
```
[リクエスト] → [ディレクトリ検証] → [スキャン実行] → [結果整理] → [レスポンス]
              ↓                    ↓            ↓
            [エラー:              [エラー:      [正常完了]
             無効ディレクトリ]      I/Oエラー]
```

## データフロー図

```
[HTTP Request]
    ↓
[HTTP Handler] (interfaces層)
    ↓
[Use Case] (application層)
    ↓
[Domain Service] (domain層)
    ↓
[Repository Implementation] (infrastructure層)
    ↓
[File System] (外部)
```

## パッケージマッピング

| ドメインエンティティ | パッケージ配置 |
|-------------------|---------------|
| FileSystemEntry | pkg/domain/entities |
| DirectoryListing | pkg/domain/entities |
| FileContent | pkg/domain/entities |
| FilePath | pkg/domain/valueobjects |
| FileSize | pkg/domain/valueobjects |
| FileSystemRepository | pkg/domain/repositories |
| DirectoryService | pkg/application/services |
| FileService | pkg/application/services |
| HTTP Handlers | pkg/interfaces/http |

## 移行時の考慮事項

1. **データ構造の互換性**: 既存のAPIレスポンス形式を完全に維持
2. **エラーハンドリング**: 既存のエラーメッセージとHTTPステータスコードを保持
3. **パフォーマンス**: メモリ使用量とレスポンス時間の維持
4. **セキュリティ**: 既存のセキュリティ制約を新しいモデルでも実装

---
*データモデル設計完了: 2025-09-20*