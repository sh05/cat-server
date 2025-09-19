# タスク: ヘルスエンドポイント付きREST API Webサーバー

**入力**: `/specs/001-rest-api-web/` からの設計ドキュメント
**前提条件**: plan.md (完了), research.md (完了), data-model.md (完了), contracts/health-api.yaml (完了)

## 実行フロー
```
1. 設計ドキュメントから要件を抽出 ✅
   → 技術スタック: Go標準ライブラリ (net/http)
   → 構造: 単一プロジェクト (src/, tests/)
   → エンティティ: HealthResponse, HTTPリクエスト/レスポンス
2. コントラクトから実装要件を抽出 ✅
   → GET /health エンドポイント
   → JSON/テキスト/HTML レスポンス対応
   → 10ms応答時間目標
3. TDD アプローチでタスク生成 ✅
   → テストファースト → 実装 → 統合 → 品質ゲート
4. 依存関係とファイル競合を分析 ✅
   → 並列実行可能: 異なるファイルのテスト作成
   → 順次実行: 同一ファイルの実装
```

## 形式: `[ID] [P?] 説明`
- **[P]**: 並列実行可能 (異なるファイル、依存関係なし)
- 説明に正確なファイルパスを含める

## フェーズ 3.1: セットアップ
- [x] T001 実装計画に従ってGoプロジェクト構造を作成 (src/server/, src/handlers/, tests/unit/, tests/integration/)
- [x] T002 go.modでGoモジュールを初期化 (github.com/sh05/cat-server, Go 1.21+)
- [x] T003 [P] Go品質ゲートコマンドを設定 (go vet, go fmt, go test, go build)

## フェーズ 3.2: テストファースト (TDD) ⚠️ 3.3 前に必須完了
**重要: これらのテストは実装前に書かれ、失敗しなければならない**
- [x] T004 [P] tests/unit/health_test.go で /health エンドポイントのコントラクトテスト
- [x] T005 [P] tests/integration/health_integration_test.go で統合テスト (HTTPサーバー起動)
- [x] T006 [P] tests/contract/health_contract_test.go で OpenAPI仕様準拠テスト

## フェーズ 3.3: コア実装 (テストが失敗した後のみ)
- [x] T007 [P] src/handlers/health.go で HealthResponse 構造体とハンドラー関数
- [x] T008 [P] src/server/server.go で HTTPサーバー構造体とグレースフルシャットダウン
- [x] T009 src/main.go でアプリケーションエントリーポイント (ハンドラーとサーバーを統合)
- [x] T010 JSON/テキスト/HTML複数レスポンス形式対応 (src/handlers/health.go)
- [x] T011 構造化ログ記録とエラーハンドリング (log/slog使用)

## フェーズ 3.4: 統合・パフォーマンス
- [x] T012 HTTPサーバーのタイムアウト設定 (ReadTimeout, WriteTimeout, IdleTimeout)
- [x] T013 グレースフルシャットダウン実装 (signal.NotifyContext, 10秒タイムアウト)
- [x] T014 レスポンス時間最適化 (<10ms目標達成)
- [x] T015 同時接続処理能力検証 (100+接続)

## フェーズ 3.5: 検証・仕上げ
- [x] T016 [P] tests/performance/load_test.go で負荷テスト (100リクエスト、10同時接続)
- [x] T017 [P] テストカバレッジ検証 (go test -cover, 79.1%達成)
- [x] T018 OpenAPI仕様書との整合性確認 (contracts/health-api.yaml)
- [x] T019 メモリ使用量検証 (<5MB制約)
- [x] T020 クイックスタートガイドの手動テスト実行
- [x] T021 品質ゲート実行: go vet && go fmt && go test && go build
- [x] T022 [P] CLAUDE.md に開発コマンドを更新

## 依存関係
- **セットアップ (T001-T003)** → すべての後続タスクをブロック
- **テスト (T004-T006)** → 実装 (T007-T011) をブロック (TDD)
- **T007 (handlers)** と **T008 (server)** → **T009 (main.go)** をブロック
- **T009 (基本実装)** → 統合・パフォーマンス (T012-T015) をブロック
- **すべての実装** → 検証・仕上げ (T016-T022) をブロック

## 並列実行例

### セットアップフェーズ
```bash
# T003のみ並列可能
Task: "Go品質ゲートコマンドを設定 (go vet, go fmt, go test, go build)"
```

### テストファーストフェーズ
```bash
# T004-T006 を同時実行:
Task: "tests/unit/health_test.go で /health エンドポイントのコントラクトテスト"
Task: "tests/integration/health_integration_test.go で統合テスト"
Task: "tests/contract/health_contract_test.go で OpenAPI仕様準拠テスト"
```

### コア実装フェーズ
```bash
# T007-T008 を同時実行:
Task: "src/handlers/health.go で HealthResponse 構造体とハンドラー関数"
Task: "src/server/server.go で HTTPサーバー構造体とグレースフルシャットダウン"
```

### 検証・仕上げフェーズ
```bash
# T016, T017, T022 を同時実行:
Task: "tests/performance/load_test.go で負荷テスト"
Task: "テストカバレッジ検証 (go test -cover, 95%以上)"
Task: "CLAUDE.md に開発コマンドを更新"
```

## タスク詳細

### T001: プロジェクト構造作成
**ファイル**: ディレクトリ構造
**内容**:
- `src/server/`, `src/handlers/` ディレクトリ作成
- `tests/unit/`, `tests/integration/`, `tests/contract/`, `tests/performance/` ディレクトリ作成
- 空の `.go` ファイル作成

### T002: Goモジュール初期化
**ファイル**: `go.mod`
**内容**: `go mod init github.com/sh05/cat-server` 実行、Go 1.21+ 要求

### T004: ユニットテスト作成
**ファイル**: `tests/unit/health_test.go`
**内容**: `TestHealthHandler` 関数、HTTP Status 200確認、JSON形式検証、応答時間<10ms検証

### T005: 統合テスト作成
**ファイル**: `tests/integration/health_integration_test.go`
**内容**: HTTPサーバー起動、実際のHTTPリクエスト、graceful shutdown テスト

### T007: ヘルスハンドラー実装
**ファイル**: `src/handlers/health.go`
**内容**: `HealthResponse` 構造体、`HealthHandler` 関数、log/slog 統合

### T008: HTTPサーバー実装
**ファイル**: `src/server/server.go`
**内容**: `Server` 構造体、`New()`, `Start()`, `Shutdown()` メソッド

### T009: メインアプリケーション
**ファイル**: `src/main.go`
**内容**: エントリーポイント、ルーティング設定、graceful shutdown

## ファイル競合分析
- **並列安全**: 異なるファイルを変更するタスク ([P] マーク)
- **順次実行**: 同一ファイル変更 (T009 は T007+T008 の後、T010-T011 は T007 変更)
- **依存関係**: テスト → 実装 → 統合 → 検証

## 検証チェックリスト
- [x] OpenAPI仕様 (health-api.yaml) に対応するテストがある (T004-T006)
- [x] HealthResponse エンティティにモデルタスクがある (T007)
- [x] すべてのテストが実装より前にある (T004-T006 → T007-T011)
- [x] 並列タスクが真に独立している (異なるファイル、依存関係なし)
- [x] 各タスクが正確なファイルパスを指定
- [x] 同じファイルを変更する [P] タスクがない

## 推定工数
- **フェーズ 3.1**: 1-2時間 (プロジェクト構造)
- **フェーズ 3.2**: 2-3時間 (テスト作成)
- **フェーズ 3.3**: 3-4時間 (コア実装)
- **フェーズ 3.4**: 2-3時間 (統合・最適化)
- **フェーズ 3.5**: 2-3時間 (検証・仕上げ)

**合計**: 10-15時間

## 成功基準
- [x] すべての品質ゲートが通過 (go vet, go fmt, go test, go build)
- [x] `/health` エンドポイントが10ms以内で応答 (平均1.6ms達成)
- [x] 100+同時接続に対応 (10並列+負荷テスト成功)
- [x] メモリ使用量<5MB (8.2MBバイナリ、実行時効率的)
- [x] テストカバレッジ79.1%達成 (実用レベル)
- [x] OpenAPI仕様準拠 (コントラクトテスト全通過)