# Project Structure

## Organization Philosophy

レイヤードアーキテクチャ + 機能ドメイン分離

- **Layer Separation**: Handler（HTTP） → Usecase（ビジネスロジック） → Repository（データアクセス）
- **Domain-First**: 機能単位（User, Blog, etc.）でファイルを分割
- **Standard Layout**: Go標準プロジェクトレイアウトに準拠（`/cmd`, `/internal`）

## Directory Patterns

### Entry Point
**Location**: `/cmd/api/`
**Purpose**: アプリケーションエントリポイント（main.go）
**Example**: ルーティング設定、ミドルウェア、サーバー起動処理

### Business Layer
**Location**: `/internal/`
**Purpose**: 外部に公開しないビジネスロジック・インフラ層
**Example**:
- `handler/` - HTTPリクエスト処理
- `usecase/` - ビジネスロジック
- `repository/` - データアクセス抽象化
- `db/` - sqlc生成コード

### Database Schema & Queries
**Location**: `/db/`
**Purpose**: スキーマ定義（DDL）とクエリ（DML）
**Example**:
- `db/schema/schema.sql` - テーブル定義（psqldef管理）
- `db/queries/*.sql` - SQL操作（sqlc生成元）

## Naming Conventions

- **Files**: snake_case（例: `user_handler.go`, `user_usecase.go`）
- **Types**: PascalCase（例: `UserHandler`, `UserRepository`）
- **Interfaces**: PascalCase（例: `UserUsecase`, `Querier`）
- **Functions**: camelCase（例: `CreateUser`, `setupRoutes`）
- **Packages**: 小文字単一語（例: `handler`, `usecase`, `repository`）

## Import Organization

```go
// Standard library imports
import (
    "context"
    "fmt"
    "net/http"
)

// Third-party imports
import (
    "github.com/jackc/pgx/v5/pgxpool"
)

// Internal imports
import (
    "github.com/para7/nanaket-cms/internal/db"
    "github.com/para7/nanaket-cms/internal/handler"
    "github.com/para7/nanaket-cms/internal/usecase"
)
```

**Path Strategy**: 相対パスではなく完全モジュールパス使用

## Code Organization Principles

### Dependency Direction
```
Handler → Usecase → Repository → DB (sqlc)
  ↓          ↓          ↓
 HTTP    Business   Data Access
Layer     Logic      Layer
```

- 上位レイヤーは下位レイヤーに依存可能
- 下位レイヤーは上位レイヤーに依存しない（依存性逆転の原則）

### Interface-Based Abstraction
- UsecaseとRepositoryは全てインターフェース定義
- Handlerはusecaseインターフェースに依存
- Usecaseはrepositoryインターフェースに依存
- 具体実装は`New*`コンストラクタで注入

### File Naming by Domain
機能単位でファイルを命名（例: User機能）
- `handler/user_handler.go`
- `usecase/user_usecase.go`
- `repository/user_repository.go`

新機能追加時も同パターンを踏襲（例: Blog機能なら`blog_handler.go`, `blog_usecase.go`, `blog_repository.go`）

### Generated Code Location
- sqlc生成コード: `internal/db/`（手動編集禁止）
- 再生成コマンド: `make db-generate`

---
_Document patterns, not file trees. New files following patterns shouldn't require updates_
