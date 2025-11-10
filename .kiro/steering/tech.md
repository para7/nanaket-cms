# Technology Stack

## Architecture

クリーンアーキテクチャベースの3層構造（Handler → Usecase → Repository）

## Core Technologies

- **Language**: Go 1.25.3
- **Database**: PostgreSQL 18 (pgx/v5ドライバ)
- **HTTP Server**: 標準ライブラリ `net/http`（フレームワークレス設計）
- **SQL Generator**: sqlc 1.30.0（型安全なクエリ生成）

## Key Libraries

- **pgxpool**: コネクションプール管理
- **pgx/v5**: PostgreSQL高性能ドライバ
- **sqlc**: SQL → Go構造体生成（コンパイル時型チェック）

## Development Standards

### Type Safety
- Goの静的型付けを最大限活用
- sqlcによるDB型とGo型の一致保証
- インターフェース指向設計（Usecase、Repositoryは全てinterface定義）

### Code Quality
- **Linter**: golangci-lint（`make lint`, `make lint-fix`）
- **Naming**: 明確な責務に基づく命名（Handler/Usecase/Repository）
- **Error Handling**: 標準エラー返却、HTTPステータスコード適切使用

### Testing
現在、テストフレームワーク未導入（今後追加予定）

## Development Environment

### Required Tools
- Go 1.25.3+
- Docker & Docker Compose
- PostgreSQL 18（Dockerで提供）
- sqlc（`go tool sqlc`）
- psqldef（スキーママイグレーション）
- golangci-lint

### Common Commands
```bash
# Dev Setup
make dev              # DB起動 + マイグレーション + sqlc生成

# Database
make db-up            # PostgreSQL起動
make db-migrate       # スキーマ適用（psqldef）
make db-generate      # sqlcコード生成
make db-reset         # DB完全リセット

# Development
make run              # アプリケーション起動（ポート8080）
make lint             # コード静的解析
make lint-fix         # 自動修正可能なlintエラーを修正
```

## Key Technical Decisions

### Why Clean Architecture?
- レイヤー間の依存を単方向に保ち、ビジネスロジックとインフラを分離
- テスト容易性の向上（各層をモック可能）

### Why Standard Library HTTP?
- 軽量で学習コスト低
- Go 1.22+の新しいルーティング機能（`GET /path/{id}`）を活用

### Why sqlc over ORM?
- SQL優先アプローチ（SQLファイルから型安全なGoコード生成）
- パフォーマンスの透明性（生成されたコードが明確）
- マイグレーション管理との分離（psqldefで宣言的スキーマ管理）

### Why pgxpool?
- Go向け最速PostgreSQLドライバ
- 本番環境に適したコネクションプール機能

---
_Document standards and patterns, not every dependency_
