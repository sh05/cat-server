# データモデル設計: ヘルスエンドポイント

**作成日**: 2025-09-20
**仕様元**: [spec.md](./spec.md)
**技術根拠**: [research.md](./research.md)

## 概要

ヘルスエンドポイント `/health` で使用されるデータ構造とHTTPリクエスト/レスポンスの定義。最小実装として、HTTP Status 200を返す基本的なヘルスチェック機能を提供。

## エンティティ定義

### 1. HealthResponse

**目的**: サーバーの稼働状況を表すレスポンスエンティティ

**属性**:
- HTTPステータスコード: 200 (固定)
- レスポンスボディ: 任意形式 (JSON、プレーンテキスト、または空)

**Go構造体表現** (実装時の参考):
```go
// HealthResponse represents the health check response
type HealthResponse struct {
    Status  string    `json:"status"`
    Time    time.Time `json:"timestamp"`
}
```

**制約・ルール**:
- HTTPステータスコードは常に200でなければならない
- レスポンスボディは任意形式で構わない
- レスポンス時間は10ms以下が目標

### 2. HTTPリクエスト構造

**エンドポイント**: `GET /health`

**リクエスト仕様**:
- HTTPメソッド: GET
- パス: `/health`
- ヘッダー: 不要
- ボディ: なし
- 認証: 不要

**例**:
```http
GET /health HTTP/1.1
Host: localhost:8080
```

### 3. HTTPレスポンス構造

**成功レスポンス**:
- HTTPステータス: 200 OK
- Content-Type: 任意 (application/json推奨)
- ボディ: 任意形式

**JSON形式の例**:
```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 58

{
    "status": "ok",
    "timestamp": "2025-09-20T10:15:30Z"
}
```

**プレーンテキスト形式の例**:
```http
HTTP/1.1 200 OK
Content-Type: text/plain
Content-Length: 2

OK
```

**空ボディの例**:
```http
HTTP/1.1 200 OK
Content-Length: 0

```

## エラーハンドリング

### サーバーエラー対応

現在の仕様では、基本的なサーバー稼働確認のみを行うため、以下のエラーパターンを考慮：

**5xx サーバーエラー**:
- サーバー内部エラーが発生した場合
- 通常、HTTP 500 Internal Server Error
- エラー詳細はログに記録、レスポンスには含めない

### ログ記録エンティティ

**構造化ログエントリ**:
```go
type HealthLogEntry struct {
    Timestamp   time.Time
    Level       string    // "INFO", "ERROR"
    Message     string
    RequestID   string    // 任意
    RemoteAddr  string
    UserAgent   string
    Duration    time.Duration
}
```

**ログ記録例**:
```json
{
    "timestamp": "2025-09-20T10:15:30Z",
    "level": "INFO",
    "message": "health check requested",
    "remote_addr": "192.168.1.100:54321",
    "user_agent": "curl/7.68.0",
    "duration": "1.2ms"
}
```

## 状態遷移

### ヘルスチェック状態

```
[リクエスト受信] → [処理実行] → [レスポンス送信]
       ↓             ↓              ↓
   ログ記録開始   サーバー状態確認   ログ記録完了
```

**状態詳細**:
1. **リクエスト受信**: HTTPリクエストの受け取り
2. **処理実行**: 基本的なサーバー稼働確認（即座に完了）
3. **レスポンス送信**: HTTP 200とボディの送信

## 検証ルール

### リクエスト検証
- HTTPメソッドがGETであること
- パスが正確に `/health` であること
- 特別な認証やパラメータは不要

### レスポンス検証
- HTTPステータスコードが200であること
- レスポンス時間が10ms以下であること
- Content-Lengthヘッダーが適切に設定されること

## 実装時の考慮事項

### パフォーマンス
- メモリ割り当てを最小化
- 同時リクエスト処理能力 100+ 接続
- ガベージコレクション負荷の最小化

### 監視・運用
- 監視システムとの統合を想定
- Kubernetes liveness probe 対応可能
- ログ集約システムとの連携

### 将来拡張性
- データベース接続チェック追加可能な設計
- 外部サービス依存関係チェック対応
- 詳細なヘルス情報（メモリ、CPU使用率）追加可能

## 関連ドキュメント

- [実装計画](./plan.md)
- [技術調査結果](./research.md)
- [API仕様書](./contracts/health-api.yaml)
- [クイックスタートガイド](./quickstart.md)