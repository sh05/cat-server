# タスク: ファイル一覧取得エンドポイント (/list)

**入力**: `/specs/004-list-get-request/` からの設計ドキュメント
**前提条件**: plan.md, research.md, data-model.md, contracts/, quickstart.md

## 実行フロー (main)
```
1. 機能ディレクトリから plan.md をロード
   → 完了: Go標準ライブラリベース、既存HTTPサーバー拡張
2. 設計ドキュメント解析完了:
   → data-model.md: DirectoryPath, FileList, APIResponse エンティティ
   → contracts/: list-endpoint.yaml OpenAPI仕様 + contract_test.go
   → research.md: os.ReadDir(), flag パッケージ、JSON レスポンス形式決定
3. カテゴリ別タスク生成完了
4. タスクルール適用: TDD順序、並列実行マーク
5. 依存関係グラフ生成完了
6. 並列実行例作成完了
7. タスク完全性検証: ✅ 全要件カバー
8. 戻り値: SUCCESS (実行準備完了)
```

## 形式: `[ID] [P?] 説明`
- **[P]**: 並列実行可能 (異なるファイル、依存関係なし)
- 説明に正確なファイルパスを含める

## フェーズ 3.1: セットアップ
- [x] T001 src/services/ ディレクトリ作成とプロジェクト構造確認
- [x] T002 [P] src/main.go に flag パッケージ追加し -dir 引数解析実装
- [x] T003 [P] Go品質ゲート確認: go vet && go fmt && go test && go build

## フェーズ 3.2: テストファースト (TDD) ⚠️ 3.3 前に必須完了
**重要: これらのテストは実装前に書かれ、失敗しなければならない**
- [x] T004 [P] specs/004-list-get-request/contracts/contract_test.go でOpenAPI契約テスト実行確認
- [x] T005 [P] tests/unit/directory_service_test.go でDirectoryServiceのユニットテスト作成
- [x] T006 [P] tests/unit/list_handler_test.go でListHandlerのユニットテスト作成
- [x] T007 [P] tests/integration/list_endpoint_test.go で /list エンドポイント統合テスト作成
- [x] T008 [P] tests/performance/list_performance_test.go でパフォーマンステスト作成

## フェーズ 3.3: コア実装 (テストが失敗した後のみ)
- [x] T009 [P] src/services/directory.go でDirectoryService実装 (data-model.mdのDirectoryPath対応)
- [x] T010 src/handlers/list.go でListHandler実装 (T009後)
- [x] T011 src/server/server.go に GET /list ルート追加 (T010後)
- [x] T012 src/main.go でディレクトリパス設定とサーバー統合 (T002,T011後)

## フェーズ 3.4: 統合
- [x] T013 src/handlers/list.go でエラーハンドリング統一 (health エンドポイントと同様)
- [x] T014 [P] src/handlers/list.go でslogを使った構造化ログ追加
- [x] T015 src/services/directory.go で入力検証強化 (パストラバーサル対策)

## フェーズ 3.5: 仕上げ
- [x] T016 [P] specs/004-list-get-request/contracts/contract_test.go で全契約テスト通過確認
- [x] T017 [P] specs/004-list-get-request/quickstart.md の実行と検証
- [x] T018 [P] README.md 更新 (新エンドポイント情報追加)
- [x] T019 品質ゲート最終確認: go vet && go fmt && go test && go build
- [x] T020 [P] プロジェクトルートでquickstart.mdシナリオ実行とデモ

## 依存関係
- セットアップ (T001-T003) がすべての前提
- テスト作成 (T004-T008) が実装 (T009-T012) より前
- T009 が T010 をブロック
- T010 が T011 をブロック
- T002,T011 が T012 をブロック
- 統合 (T013-T015) が仕上げ (T016-T020) より前

## 並列実行例
```bash
# フェーズ 3.1: セットアップ (並列実行)
Task: "src/main.go に flag パッケージ追加し -dir 引数解析実装"
Task: "Go品質ゲート確認: go vet && go fmt && go test && go build"

# フェーズ 3.2: テスト作成 (全て並列実行可能)
Task: "specs/004-list-get-request/contracts/contract_test.go でOpenAPI契約テスト実行確認"
Task: "tests/unit/directory_service_test.go でDirectoryServiceのユニットテスト作成"
Task: "tests/unit/list_handler_test.go でListHandlerのユニットテスト作成"
Task: "tests/integration/list_endpoint_test.go で /list エンドポイント統合テスト作成"
Task: "tests/performance/list_performance_test.go でパフォーマンステスト作成"

# フェーズ 3.5: 仕上げ (一部並列実行可能)
Task: "specs/004-list-get-request/contracts/contract_test.go で全契約テスト通過確認"
Task: "specs/004-list-get-request/quickstart.md の実行と検証"
Task: "README.md 更新 (新エンドポイント情報追加)"
Task: "プロジェクトルートでquickstart.mdシナリオ実行とデモ"
```

## 詳細タスク仕様

### T001: プロジェクト構造確認
**場所**: `src/services/`
**成果物**: services ディレクトリ作成
**検証**: `ls src/services/` でディレクトリ存在確認

### T002: コマンドライン引数実装 [P]
**場所**: `src/main.go`
**成果物**: flag パッケージ使用、-dir 引数解析
**検証**: `go run src/main.go -dir ./test` でカスタムディレクトリ指定可能
**実装**:
```go
var dirFlag = flag.String("dir", "./files/", "Directory to list files from")
flag.Parse()
```

### T003: 品質ゲート確認 [P]
**場所**: プロジェクトルート
**検証**: すべてのコマンドが成功
```bash
go vet ./...
go fmt ./...
go test ./...
go build ./...
```

### T004: 契約テスト実行確認 [P]
**場所**: `specs/004-list-get-request/contracts/contract_test.go`
**成果物**: 既存契約テストの実行確認
**検証**: `go test ./specs/004-list-get-request/contracts/ -v` で適切に失敗

### T005: DirectoryService ユニットテスト [P]
**場所**: `tests/unit/directory_service_test.go`
**成果物**: DirectoryService の完全なテスト
**テストケース**:
- 有効ディレクトリのファイル一覧取得
- 隠しファイル(.から始まるファイル)含有確認
- 存在しないディレクトリエラー
- 権限エラーハンドリング
- 空ディレクトリ処理

### T006: ListHandler ユニットテスト [P]
**場所**: `tests/unit/list_handler_test.go`
**成果物**: ListHandler のHTTPテスト
**テストケース**:
- 正常なGETリクエスト
- JSON レスポンス形式確認
- Content-Type: application/json 確認
- エラーレスポンス形式
- HTTP メソッド制限 (POSTなど禁止)

### T007: 統合テスト [P]
**場所**: `tests/integration/list_endpoint_test.go`
**成果物**: エンドツーエンドテスト
**テストケース**:
- httptest.Server 使用したフル統合テスト
- 実際のファイルシステムとの統合
- quickstart.md シナリオの自動化

### T008: パフォーマンステスト [P]
**場所**: `tests/performance/list_performance_test.go`
**成果物**: 負荷・応答時間テスト
**検証目標**:
- 1000ファイル以下で応答時間 <100ms
- 同時リクエスト処理能力
- メモリ使用量測定

### T009: DirectoryService実装 [P]
**場所**: `src/services/directory.go`
**成果物**: DirectoryService 構造体と関数
**実装内容**:
```go
package services

import (
    "os"
    "path/filepath"
)

type DirectoryService struct {
    basePath string
}

func NewDirectoryService(path string) *DirectoryService
func (ds *DirectoryService) ListFiles() ([]string, error)
func (ds *DirectoryService) ValidatePath() error
```

### T010: ListHandler実装
**場所**: `src/handlers/list.go`
**成果物**: ListHandler 関数
**依存**: T009 (DirectoryService)
**実装内容**:
```go
package handlers

import (
    "encoding/json"
    "net/http"
    "github.com/sh05/cat-server/src/services"
)

func ListHandler(directoryService *services.DirectoryService) http.HandlerFunc
```

### T011: ルート追加
**場所**: `src/server/server.go`
**成果物**: GET /list ルート登録
**依存**: T010 (ListHandler)
**実装**: `mux.HandleFunc("GET /list", handlers.ListHandler(directoryService))`

### T012: main.go統合
**場所**: `src/main.go`
**成果物**: フラグ解析からサーバー起動まで統合
**依存**: T002, T011
**実装**: フラグパース、DirectoryService作成、サーバー設定

### T013: エラーハンドリング統一
**場所**: `src/handlers/list.go`
**成果物**: health エンドポイントと一貫したエラーレスポンス
**実装**: 400/403/500 エラーの統一JSON形式

### T014: 構造化ログ追加 [P]
**場所**: `src/handlers/list.go`
**成果物**: slog使用したリクエスト/レスポンスログ
**実装**: health エンドポイントと同様のログ形式

### T015: セキュリティ強化
**場所**: `src/services/directory.go`
**成果物**: パストラバーサル攻撃対策、入力検証
**実装**: `../`チェック、nullバイト検証、パス正規化

### T016-T020: 仕上げフェーズ
**成果物**:
- 全契約テスト通過
- quickstart.md実行成功
- README.md更新
- 品質ゲート合格
- デモンストレーション完了

## 注意事項
- **TDD必須**: テスト (T004-T008) が実装 (T009-T012) より前
- **並列制約**: 同じファイルを変更するタスクは順次実行
- **品質ゲート**: 各フェーズ完了後に実行必須
- **憲法準拠**: Go標準ライブラリのみ、外部依存関係なし

## 検証チェックリスト
- [x] すべての契約にテストがある (T004: OpenAPI契約テスト)
- [x] すべてのエンティティにサービスがある (T009: DirectoryService)
- [x] すべてのテストが実装より前 (T004-T008 → T009-T012)
- [x] 並列タスクが独立している ([P]マーク済み)
- [x] 各タスクが正確なファイルパス指定
- [x] 同じファイルの [P] タスクなし

---
**実行準備完了**: 全20タスク、TDD順序、並列実行最適化済み