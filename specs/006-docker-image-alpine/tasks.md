# タスク: Docker Image作成機能

**入力**: `/specs/006-docker-image-alpine/` からの設計ドキュメント
**前提条件**: plan.md (必須), research.md, data-model.md, contracts/

## 実行フロー (main)
```
1. 機能ディレクトリから plan.md をロード
   → 技術スタック: Alpine Linux, Docker, マルチステージビルド
   → 成果物: Dockerfileのみ
2. オプションの設計ドキュメントをロード:
   → data-model.md: Dockerイメージ、ビルドステージ、実行環境エンティティ
   → contracts/: docker-build.yaml → Docker操作テスト
   → research.md: セキュリティ、パフォーマンス、デプロイメント決定
3. カテゴリ別にタスクを生成:
   → セットアップ: .dockerignore, テスト環境
   → テスト: Docker契約テスト、統合テスト
   → コア: Dockerfile、ビルドスクリプト
   → 統合: Docker統合テスト
   → 仕上げ: パフォーマンステスト、ドキュメント更新
4. タスクルールを適用:
   → 異なるファイル = 並列に [P] をマーク
   → TDD原則: テスト → 実装 → 検証
5. タスクを順番に番号付け (T001-T011)
6. 依存関係: セットアップ → テスト → 実装 → 統合 → 仕上げ
7. 並列実行: .dockerignore と契約テストは独立
8. タスクの完全性を検証: ✅ すべてのコントラクト、エンティティ、シナリオを網羅
9. 戻り値: SUCCESS (実行準備完了)
```

## 形式: `[ID] [P?] 説明`
- **[P]**: 並列実行可能 (異なるファイル、依存関係なし)
- 説明に正確なファイルパスを含める

## パス規約
- **単一プロジェクト**: リポジトリルートに Dockerfile, .dockerignore
- **テスト**: `tests/docker/` 新規作成
- 既存の `src/`, `tests/` 構造を維持

## フェーズ 3.1: セットアップ
- [x] T001 [P] .dockerignore ファイルを作成してビルド最適化 (./dockerignore)
- [x] T002 [P] Docker テストディレクトリの作成 (tests/docker/)

## フェーズ 3.2: テストファースト (TDD) ⚠️ 3.3 前に必須完了
- [x] T003 [P] Docker ビルド契約テストの作成 (tests/docker/build_test.go)
- [x] T004 [P] Docker 実行契約テストの作成 (tests/docker/run_test.go)
- [x] T005 [P] Docker イメージ検査テストの作成 (tests/docker/inspect_test.go)

## フェーズ 3.3: コア実装
- [x] T006 マルチステージ Dockerfile の作成 (./Dockerfile)
- [x] T007 Docker ビルド・実行スクリプトの作成 (scripts/docker-build.sh)

## フェーズ 3.4: 統合とテスト
- [x] T008 Docker統合テストシナリオの実装 (tests/docker/integration_test.go)
- [x] T009 Docker セキュリティテストの実装 (tests/docker/security_test.go)

## フェーズ 3.5: 仕上げと最適化
- [x] T010 [P] Docker パフォーマンステストの実装 (tests/docker/performance_test.go)
- [x] T011 [P] CLAUDE.md ドキュメントの更新 (Docker コマンド追加済み)

---

## 依存関係グラフ
```
T001,T002 (並列) → T003,T004,T005 (並列) → T006 → T007 → T008,T009 (並列) → T010,T011 (並列)
```

## 並列実行例

### 第1波: セットアップタスク
```bash
# 並列実行可能
Task agent: "T001 - .dockerignore ファイルを作成してビルド最適化"
Task agent: "T002 - Docker テストディレクトリの作成"
```

### 第2波: テストファースト実装
```bash
# 並列実行可能 (異なるテストファイル)
Task agent: "T003 - Docker ビルド契約テストの作成"
Task agent: "T004 - Docker 実行契約テストの作成"
Task agent: "T005 - Docker イメージ検査テストの作成"
```

### 第3波: コア実装 (順次実行)
```bash
# T006 → T007 の順序で実行 (Dockerfileが前提)
Task agent: "T006 - マルチステージ Dockerfile の作成"
Task agent: "T007 - Docker ビルド・実行スクリプトの作成"
```

### 第4波: 統合テスト
```bash
# 並列実行可能
Task agent: "T008 - Docker統合テストシナリオの実装"
Task agent: "T009 - Docker セキュリティテストの実装"
```

### 第5波: 仕上げ
```bash
# 並列実行可能
Task agent: "T010 - Docker パフォーマンステストの実装"
Task agent: "T011 - CLAUDE.md ドキュメントの更新"
```

---

## 詳細タスク仕様

### T001: .dockerignore ファイルを作成してビルド最適化
**ファイル**: `./dockerignore`
**目的**: Dockerビルド時の不要ファイル除外によるビルド速度向上
**要件**:
- .git/, node_modules/, .DS_Store の除外
- tests/, specs/, .specify/ の除外
- bin/, build/ ディレクトリの除外
- 一時ファイルとログファイルの除外

### T002: Docker テストディレクトリの作成
**ディレクトリ**: `tests/docker/`
**目的**: Docker関連テストの整理された配置
**要件**:
- Go モジュール構造に準拠
- 共通テストヘルパーファイルの準備
- テストデータディレクトリの作成

### T003: Docker ビルド契約テストの作成
**ファイル**: `tests/docker/build_test.go`
**目的**: docker build 操作の契約仕様テスト
**要件**:
- ビルド成功シナリオのテスト
- イメージサイズ < 50MB の検証
- ビルド時間 < 60秒 の検証
- エラーハンドリングのテスト

### T004: Docker 実行契約テストの作成
**ファイル**: `tests/docker/run_test.go`
**目的**: docker run 操作の契約仕様テスト
**要件**:
- コンテナ起動成功シナリオのテスト
- 起動時間 < 2秒 の検証
- ポート公開の検証
- ヘルスチェック応答のテスト

### T005: Docker イメージ検査テストの作成
**ファイル**: `tests/docker/inspect_test.go`
**目的**: イメージ構造と設定の検証
**要件**:
- 非rootユーザー(app)での実行確認
- 公開ポート 8080/tcp の確認
- Alpine Linux ベースの確認
- セキュリティ脆弱性チェック

### T006: マルチステージ Dockerfile の作成
**ファイル**: `./Dockerfile`
**目的**: Alpine ベースの軽量Docker イメージ作成
**要件**:
- ステージ1: golang:alpine でのビルド環境
- ステージ2: alpine:latest での実行環境
- CGO_ENABLED=0 での静的リンクビルド
- 非rootユーザー(app)での実行
- ヘルスチェック設定
- research.md の決定事項を反映

### T007: Docker ビルド・実行スクリプトの作成
**ファイル**: `scripts/docker-build.sh`
**目的**: 開発者向けの便利なビルド・実行スクリプト
**要件**:
- docker build の実行
- イメージサイズの表示
- docker run での起動
- ログ出力とクリーンアップ

### T008: Docker統合テストシナリオの実装
**ファイル**: `tests/docker/integration_test.go`
**目的**: エンドツーエンドのDocker動作テスト
**要件**:
- quickstart.md のシナリオ実装
- API エンドポイントへのアクセステスト
- ボリュームマウントテスト
- 環境変数テスト

### T009: Docker セキュリティテストの実装
**ファイル**: `tests/docker/security_test.go`
**目的**: コンテナセキュリティの検証
**要件**:
- 非rootユーザー実行の確認
- プロセス権限の検査
- ファイルシステム権限の検証
- 不要なパッケージの非存在確認

### T010: Docker パフォーマンステストの実装
**ファイル**: `tests/docker/performance_test.go`
**目的**: パフォーマンス目標の検証
**要件**:
- イメージサイズ < 50MB の測定
- 起動時間 < 2秒 の測定
- メモリ使用量の監視
- レスポンス時間の測定

### T011: CLAUDE.md ドキュメントの更新
**ファイル**: `CLAUDE.md`
**目的**: Docker 開発コマンドの追加 (既に完了済み)
**要件**:
- Dockerビルドコマンドの追加
- コンテナ管理コマンドの追加
- デバッグとトラブルシューティング手順
- パフォーマンス監視コマンド

---

## 品質ゲート要件
各タスク完了時に以下を確認:
- [ ] go vet: 静的解析エラーなし
- [ ] go fmt: コードフォーマット済み
- [ ] go test: 全テストパス
- [ ] go build: コンパイル成功
- [ ] docker build: イメージビルド成功 (T006以降)

## 憲法準拠チェック
- [x] **Go言語標準**: Docker関連もGoテストで実装
- [x] **テスト要件**: 全機能にテストを先行作成 (TDD)
- [x] **品質ゲート**: 各段階で品質確認
- [x] **簡素性**: 最小限の成果物 (Dockerfileと関連ファイルのみ)

## 成功基準
- [x] Alpine ベースの軽量イメージ (<50MB)
- [x] マルチステージビルドの実装
- [x] 非rootユーザーでの実行
- [x] 高速起動 (<2秒)
- [x] 包括的なテストカバレッジ
- [x] セキュリティベストプラクティス準拠