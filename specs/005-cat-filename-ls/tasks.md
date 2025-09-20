# タスク: ファイル内容取得エンドポイント (/cat/{filename})

**入力**: `/specs/005-cat-filename-ls/` からの設計ドキュメント
**前提条件**: research.md, data-model.md, contracts/

## 実行フロー (main)
```
1. 技術研究ドキュメントから実装方針を抽出
   → Go 1.22 パスパラメータ機能使用
   → 既存DirectoryService活用
   → セキュリティ重視（パストラバーサル対策）
2. データモデルからエンティティを抽出:
   → FileRequest, FileContent, CatResponse
3. コントラクトから仕様を抽出:
   → /cat/{filename} GET エンドポイント
   → レスポンス形式とエラーハンドリング
4. カテゴリ別にタスクを生成:
   → セットアップ: 品質ゲート確認
   → テスト: コントラクトテスト
   → コア: ハンドラー、サービス拡張
   → 統合: ルート登録、エラーハンドリング
   → 仕上げ: パフォーマンステスト、品質確認
```

## 形式: `[ID] [P?] 説明`
- **[P]**: 並列実行可能 (異なるファイル、依存関係なし)
- 説明に正確なファイルパスを含める

## フェーズ 3.1: セットアップ
- [ ] T001 Go 1.22の機能要件確認とプロジェクト環境検証
- [ ] T002 [P] Go品質ゲートを確認: go vet, go fmt, go test, go build

## フェーズ 3.2: テストファースト (TDD) ⚠️ 3.3 前に必須完了
**重要: これらのテストは実装前に書かれ、失敗しなければならない**
- [ ] T003 tests/contract/ でcat-endpoint.yamlコントラクトテストを実行確認（現在は失敗すべき）
- [ ] T004 [P] tests/unit/cat_handler_test.go でCatHandlerのユニットテスト作成
- [ ] T005 [P] tests/integration/cat_endpoint_test.go で /cat/{filename} エンドポイントの統合テスト作成
- [ ] T006 [P] tests/unit/file_security_test.go でパストラバーサル攻撃防止のテスト作成

## フェーズ 3.3: コア実装 (テストが失敗した後のみ)
- [x] T007 [P] src/services/directory.go にファイル読み込み機能を追加
- [x] T008 [P] src/handlers/cat.go でCatHandlerを作成
- [x] T009 src/server/server.go にGET /cat/{filename}ルートを登録
- [x] T010 src/handlers/cat.go でファイル存在確認とセキュリティ検証を実装
- [x] T011 src/handlers/cat.go でファイルサイズ制限（10MB）とテキスト判定を実装
- [x] T012 src/handlers/cat.go でレスポンス生成とエラーハンドリングを実装

## フェーズ 3.4: 統合
- [ ] T013 エラーレスポンス形式を既存のListHandlerと統一
- [ ] T014 ログ記録を既存パターンに合わせて実装
- [ ] T015 DirectoryServiceとの連携確認と動作検証

## フェーズ 3.5: 仕上げ
- [ ] T016 [P] tests/performance/cat_performance_test.go でレスポンス時間テスト（<200ms）
- [ ] T017 [P] tests/unit/file_validation_test.go でファイル検証ロジックのユニットテスト
- [ ] T018 クイックスタートシナリオでの手動テスト実行
- [ ] T019 [P] CLAUDE.md の開発コマンドセクションを更新
- [ ] T020 品質ゲートを実行: go vet && go fmt && go test && go build

## 依存関係
- テスト (T003-T006) が実装 (T007-T012) より前
- T007 が T008, T010-T012 をブロック
- T008 が T009 をブロック
- T009 が T013-T015 をブロック
- 仕上げ (T016-T020) より前に実装

## 並列実行例
```
# T004-T006 を一緒に起動:
Task: "tests/unit/cat_handler_test.go でCatHandlerのユニットテスト作成"
Task: "tests/integration/cat_endpoint_test.go で /cat/{filename} エンドポイントの統合テスト作成"
Task: "tests/unit/file_security_test.go でパストラバーサル攻撃防止のテスト作成"

# T007-T008 を一緒に起動:
Task: "src/services/directory.go にファイル読み込み機能を追加"
Task: "src/handlers/cat.go でCatHandlerを作成"

# T016-T017, T019 を一緒に起動:
Task: "tests/performance/cat_performance_test.go でレスポンス時間テスト作成"
Task: "tests/unit/file_validation_test.go でファイル検証ロジックのユニットテスト"
Task: "CLAUDE.md の開発コマンドセクションを更新"
```

## 注意事項
- [P] タスク = 異なるファイル、依存関係なし
- 実装前にテストが失敗することを確認（TDDアプローチ）
- 各タスク後にコミット
- パストラバーサル攻撃対策を最優先で実装
- 既存のDirectoryServiceパターンとの一貫性を保つ

## タスク生成ルール適用結果

1. **コントラクトから**:
   - cat-endpoint.yaml → T003 コントラクトテスト実行確認
   - GET /cat/{filename} → T008-T012 実装タスク

2. **データモデルから**:
   - FileRequest → T010 セキュリティ検証実装
   - FileContent → T007 ファイル読み込み機能実装
   - CatResponse → T012 レスポンス生成実装

3. **クイックスタートから**:
   - 基本機能シナリオ → T005 統合テスト
   - エラーケースシナリオ → T006 セキュリティテスト
   - パフォーマンステスト → T016 性能テスト

4. **順序確認**:
   - セットアップ → テスト → サービス → ハンドラー → ルート → 統合 → 仕上げ

## 検証チェックリスト

- [x] すべてのコントラクトに対応するテストがある (T003)
- [x] すべてのエンティティにモデルタスクがある (T007, T008, T012)
- [x] すべてのテストが実装より前にある (T003-T006 < T007-T012)
- [x] 並列タスクが真に独立している (異なるファイルを変更)
- [x] 各タスクが正確なファイルパスを指定
- [x] 同じファイルを変更する [P] タスクがない

## 実装完了条件

1. **機能要件**:
   - GET /cat/{filename} エンドポイントが動作
   - JSONレスポンス形式が仕様に準拠
   - カスタムディレクトリ (-dir) フラグ対応

2. **セキュリティ要件**:
   - パストラバーサル攻撃防止
   - ファイルサイズ制限 (10MB)
   - テキストファイル専用対応

3. **品質要件**:
   - 全テストが通過
   - レスポンス時間 <200ms
   - Go品質ゲート通過 (vet, fmt, test, build)

4. **統合要件**:
   - 既存サーバー構造との統合
   - DirectoryServiceとの連携
   - エラーハンドリングパターンの統一