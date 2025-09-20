# 技術調査: Docker Image作成機能

**フェーズ0出力** | **日付**: 2025-09-20

## 調査概要
Docker Image作成機能における3つの要明確化項目（セキュリティ、パフォーマンス、デプロイメント要件）について技術調査を実施し、実装方針を決定する。

## 調査結果

### 1. セキュリティ要件
**調査内容**: Alpine Linuxセキュリティベストプラクティス

**決定**: 非rootユーザーでの実行とセキュリティ強化
**根拠**:
- Alpine Linuxは最小限のパッケージ構成で攻撃面を削減
- 非rootユーザー（app）での実行により権限昇格攻撃を防止
- distroless的アプローチでランタイム依存関係を最小化

**検討した代替案**:
- scratch イメージ: 静的バイナリ作成が複雑
- ubuntu:slim: イメージサイズが大きい（約65MB vs Alpine約5MB）
- distroless/static: cgoが必要な場合に制約

**実装方針**:
- RUN adduser -D -H -s /sbin/nologin app での非rootユーザー作成
- USER app での実行ユーザー切り替え
- COPY --chown=app:app での適切な権限設定

### 2. パフォーマンス要件
**調査内容**: Goアプリケーションのマルチステージビルド最適化

**決定**: 軽量イメージ（<50MB）と高速起動（<2秒）を達成
**根拠**:
- マルチステージビルドでGoビルド環境と実行環境を分離
- 静的リンクビルド（CGO_ENABLED=0）でAlpineの互換性問題を回避
- バイナリのみの実行環境で最小リソース使用量を実現

**検討した代替案**:
- 単一ステージビルド: 開発ツールが残りサイズが大きい（約200MB）
- 動的リンクバイナリ: Alpine musl libcとの互換性問題

**実装方針**:
- ステージ1: golang:alpine でのビルド環境
- ステージ2: alpine:latest での実行環境
- CGO_ENABLED=0 GOOS=linux での静的バイナリ作成
- go build -ldflags="-w -s" でのバイナリサイズ最適化

### 3. デプロイメント要件
**調査内容**: Docker Hub/レジストリ連携なしの単純ビルド

**決定**: ローカルビルド中心のシンプル構成
**根拠**:
- 明確な成果物はDockerfileのみとの要求
- CI/CD統合は将来的な拡張として保留
- 開発者ローカル環境での即座の利用を優先

**検討した代替案**:
- GitHub Actions自動ビルド: 複雑性増加、要求範囲外
- Multi-platform build: ARM/AMD64対応は将来拡張

**実装方針**:
- docker build コマンドでの単純ビルド
- docker run での動作確認手順
- .dockerignore での不要ファイル除外

## Docker Healthcheck実装
**決定**: HTTP Healthcheckの実装
**根拠**: cat-serverはHTTPサーバーとして動作するため
**実装方針**:
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
```

## 最終アーキテクチャ決定

### Dockerイメージ構造
```
Stage 1 (Builder): golang:alpine
├── Go開発環境の設定
├── 依存関係のダウンロード
├── 静的バイナリの作成
└── バイナリの最適化

Stage 2 (Runtime): alpine:latest
├── 非rootユーザーの作成
├── 必要な実行時依存関係の追加
├── バイナリのコピー
├── ポート公開とHealthcheck
└── 非rootユーザーでの実行
```

### パフォーマンス目標
- **イメージサイズ**: < 50MB (目標 20-30MB)
- **起動時間**: < 2秒
- **ビルド時間**: < 60秒 (初回), < 30秒 (キャッシュ利用)

### セキュリティ目標
- 非rootユーザーでの実行
- 最小限の実行時依存関係
- セキュリティスキャンでの高評価

## 次ステップ
すべての要明確化項目が解決され、フェーズ1（設計とコントラクト）に進行可能。