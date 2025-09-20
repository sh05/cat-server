# タスク: Go標準ディレクトリ構造リファクタリング

**入力**: `/specs/007-go-cmd-pkg/` からの設計ドキュメント
**前提条件**: plan.md (✓), research.md (✓), data-model.md (✓), contracts/ (✓), quickstart.md (✓)

## 実行フロー (main)
```
1. 機能ディレクトリから plan.md をロード ✓
   → 抽出: Clean Architecture + DDD, cmd/pkg構造, Go標準ライブラリ
2. オプションの設計ドキュメントをロード ✓:
   → data-model.md: FileSystemEntry, DirectoryListing, FileContent エンティティ
   → contracts/: refactoring-contract.yaml → リファクタリング検証テスト
   → research.md: Strangler Fig パターン、手動DI、段階的移行
3. カテゴリ別にタスクを生成:
   → セットアップ: Go標準構造作成、依存関係確認
   → テスト: リファクタリング検証、回帰テスト
   → コア: Domain層、Infrastructure層、Application層実装
   → 統合: 依存関係注入、エンドポイント接続
   → 仕上げ: 旧コード削除、品質ゲート確認
4. Strangler Fig パターンによる段階的移行
5. タスクを順番に番号付け (T001, T002...)
6. 既存機能完全維持を最優先
7. 戻り値: SUCCESS (リファクタリング実行準備完了)
```

## 形式: `[ID] [P?] 説明`
- **[P]**: 並列実行可能 (異なるファイル、依存関係なし)
- 説明に正確なファイルパスを含める

## パス規約
- **新構造**: `cmd/cat-server/`, `pkg/domain/`, `pkg/application/`, `pkg/infrastructure/`, `pkg/interfaces/`, `internal/`
- **旧構造**: `src/` (最終段階で削除)
- **テスト**: 既存 `tests/` 構造を維持

## フェーズ 3.1: プロジェクト構造セットアップ
- [x] T001 新しいGo標準ディレクトリ構造を作成 (cmd/, pkg/, internal/)
- [x] T002 go.mod の整合性確認とモジュールパス検証
- [x] T003 [P] 既存品質ゲートの動作確認: go vet, go fmt, go test, go build

## フェーズ 3.2: テストファースト実装 (TDD) ⚠️ 3.3 前に必須完了
**重要: これらのテストは新構造用に書かれ、最初は失敗しなければならない**
- [x] T004 [P] specs/007-go-cmd-pkg/contracts/refactoring_test.go でリファクタリング検証コントラクトテスト
- [x] T005 [P] pkg/domain/entities/filesystem_entry_test.go でFileSystemEntryのユニットテスト
- [x] T006 [P] pkg/domain/entities/directory_listing_test.go でDirectoryListingのユニットテスト
- [x] T007 [P] pkg/domain/entities/file_content_test.go でFileContentのユニットテスト
- [x] T008 [P] pkg/domain/valueobjects/file_path_test.go でFilePathのユニットテスト
- [x] T009 [P] tests/integration/refactored_api_test.go で新構造でのAPI統合テスト

## フェーズ 3.3: Domain層実装 (テストが失敗した後のみ)
- [x] T010 [P] pkg/domain/entities/filesystem_entry.go でFileSystemEntryエンティティ
- [x] T011 [P] pkg/domain/entities/directory_listing.go でDirectoryListingエンティティ
- [x] T012 [P] pkg/domain/entities/file_content.go でFileContentエンティティ
- [x] T013 [P] pkg/domain/valueobjects/file_path.go でFilePathバリューオブジェクト
- [x] T014 [P] pkg/domain/valueobjects/file_size.go でFileSizeバリューオブジェクト
- [x] T015 [P] pkg/domain/repositories/filesystem_repository.go でFileSystemRepositoryインターフェース
- [x] T016 [P] pkg/domain/repositories/health_repository.go でHealthRepositoryインターフェース

## フェーズ 3.4: Infrastructure層実装
- [x] T017 [P] pkg/infrastructure/filesystem/filesystem_repository_impl.go でファイルシステム実装
- [x] T018 [P] pkg/infrastructure/logging/logger.go でログ機能実装
- [x] T019 [P] pkg/infrastructure/http/server.go でHTTPサーバー実装
- [x] T020 [P] internal/config/config.go で設定管理

## フェーズ 3.5: Application層実装
- [x] T021 [P] pkg/application/services/directory_service.go でDirectoryServiceユースケース
- [x] T022 [P] pkg/application/services/file_service.go でFileServiceユースケース
- [x] T023 [P] pkg/application/services/health_service.go でHealthServiceユースケース

## フェーズ 3.6: Interfaces層実装
- [ ] T024 pkg/interfaces/http/health_handler.go でヘルスチェックハンドラー
- [ ] T025 pkg/interfaces/http/list_handler.go でファイルリストハンドラー
- [ ] T026 pkg/interfaces/http/cat_handler.go でファイル内容ハンドラー
- [ ] T027 pkg/interfaces/http/middleware.go で共通ミドルウェア

## フェーズ 3.7: エントリーポイントと統合
- [ ] T028 cmd/cat-server/main.go で新しいメインエントリーポイント作成
- [ ] T029 cmd/cat-server/wire.go で依存関係注入ワイヤリング
- [ ] T030 新構造での動作確認とテスト実行

## フェーズ 3.8: 移行検証とクリーンアップ
- [ ] T031 既存API互換性の全面的確認 (ヘルス、リスト、cat)
- [ ] T032 Docker環境での新構造動作確認
- [ ] T033 パフォーマンステスト実行（応答時間・メモリ使用量維持確認）
- [ ] T034 旧 src/ ディレクトリの完全削除
- [ ] T035 [P] 既存テストのインポートパス更新 (tests/unit/, tests/integration/, tests/contract/)

## フェーズ 3.9: 最終品質確認
- [ ] T036 go vet, go fmt, go test, go build の全品質ゲート実行
- [ ] T037 [P] Dockerfile とスクリプトの動作確認
- [ ] T038 [P] CLAUDE.md のコマンド例更新
- [ ] T039 specs/007-go-cmd-pkg/quickstart.md の実行テスト
- [ ] T040 リファクタリング完了の最終検証

## 依存関係

### 必須順序
```
T001-T003 (セットアップ)
    ↓
T004-T009 (テスト作成)
    ↓
T010-T016 (Domain層) → T017-T020 (Infrastructure層) → T021-T023 (Application層)
    ↓
T024-T027 (Interfaces層)
    ↓
T028-T030 (統合)
    ↓
T031-T035 (検証・移行)
    ↓
T036-T040 (最終確認)
```

### 並列実行可能グループ
**グループA (Domain層)**: T010, T011, T012, T013, T014, T015, T016
**グループB (Infrastructure層)**: T017, T018, T019, T020
**グループC (Application層)**: T021, T022, T023
**グループD (テスト作成)**: T004, T005, T006, T007, T008, T009
**グループE (最終確認)**: T037, T038, T039

## 並列実行例

### パターン1: 高速開発（4並列）
```bash
# グループD: テスト作成
T004, T005, T006, T007 を並列実行

# グループA: Domain層
T010, T011, T012, T013 を並列実行

# グループB+C: サービス層
T017, T018, T021, T022 を並列実行
```

### パターン2: 段階的実装（2並列）
```bash
# フェーズ1: 基盤
T001 → (T004, T005) → (T010, T011)

# フェーズ2: 機能実装
(T017, T021) → (T024, T025)

# フェーズ3: 統合
T028 → T031 → T034
```

## 重要な制約

### 既存機能維持
- **絶対条件**: 全既存APIが完全に同じ動作をする
- **検証手段**: 既存テストスイートが100%成功
- **Docker互換性**: コンテナビルド・実行が正常動作

### 品質ゲート
- **必須**: 各フェーズ後に go vet, go fmt, go test, go build 実行
- **失敗時**: 次フェーズに進まずに修正

### セキュリティ
- **パストラバーサル防止**: 既存セキュリティ機能の完全移植
- **権限管理**: 非rootユーザー実行の維持

## 成功基準

1. **構造準拠**: cmd/, pkg/, internal/ 構造が正しく配置されている
2. **機能維持**: 全てのAPIエンドポイントが既存と同一動作
3. **テスト成功**: 既存テスト + 新規テストが全て成功
4. **品質ゲート**: go vet, go fmt, go test, go build が全て成功
5. **Docker動作**: コンテナビルド・実行が正常
6. **完全移行**: 旧 src/ ディレクトリが完全に削除されている
7. **パフォーマンス**: 応答時間・メモリ使用量が同等以下

---

**推定工数**: 40タスク, 約1-2日間
**最終更新**: 2025-09-20