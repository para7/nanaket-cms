# API開発ガイド

このドキュメントは、nanaket-cmsプロジェクトで新しいAPIエンドポイントを追加する際の標準手順と設計パターンをまとめたものです。

## 目次

1. [アーキテクチャ概要](#アーキテクチャ概要)
2. [新しいAPI追加の全体フロー](#新しいapi追加の全体フロー)
3. [各層の役割と責務](#各層の役割と責務)
4. [編集するファイルと順序](#編集するファイルと順序)
5. [実装例：記事管理API](#実装例記事管理api)
6. [命名規約](#命名規約)
7. [開発コマンド](#開発コマンド)
8. [チェックリスト](#チェックリスト)

---

## アーキテクチャ概要

本プロジェクトは**クリーンアーキテクチャ**に基づいた層構造を採用しています。

### 層構造と依存関係

```
HTTP Request
    ↓
┌─────────────────────────────────────┐
│ Handler Layer (internal/handler)    │ ← HTTPリクエスト処理、バリデーション
│ - JSON encode/decode                │
│ - HTTPステータスコード設定          │
└─────────────────────────────────────┘
    ↓ （Usecaseインターフェースに依存）
┌─────────────────────────────────────┐
│ Usecase Layer (internal/usecase)    │ ← ビジネスロジック
│ - データ変換・加工                  │
│ - エラーハンドリング                │
└─────────────────────────────────────┘
    ↓ （Repositoryインターフェースに依存）
┌─────────────────────────────────────┐
│ Repository Layer (internal/repository) │ ← データアクセス抽象化
│ - sqlc生成コードのラッピング        │
│ - クエリ実行                        │
└─────────────────────────────────────┘
    ↓
┌─────────────────────────────────────┐
│ DB Layer (internal/db)              │ ← sqlc自動生成
│ - 型安全なクエリ実行                │
└─────────────────────────────────────┘
    ↓
[PostgreSQL Database]
```

### プロジェクト構造

```
nanaket-cms/
├── cmd/
│   └── api/
│       └── main.go              # エントリポイント、ルーティング設定
├── internal/
│   ├── handler/                 # HTTPハンドラー層
│   │   └── user_handler.go
│   ├── usecase/                 # ビジネスロジック層
│   │   └── user_usecase.go
│   ├── repository/              # データアクセス層
│   │   └── user_repository.go
│   └── db/                      # sqlc生成コード（make db-generateで生成）
├── db/
│   ├── schema/
│   │   ├── schema.sql           # テーブル定義
│   │   └── functions.sql        # DB関数・トリガー
│   └── queries/
│       └── users.sql            # SQLクエリ定義
├── sqlc.yaml                    # sqlc設定ファイル
├── Makefile                     # 開発用コマンド
└── docker-compose.yml           # PostgreSQL環境
```

---

## 新しいAPI追加の全体フロー

新しいAPIを追加する際は、以下の順序で実装を進めます。

### 1. データベース層の準備

#### 1-1. スキーマ定義の追加

**ファイル**: `db/schema/schema.sql`

```sql
-- 新しいテーブルを追加
CREATE TABLE IF NOT EXISTS articles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    published_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- インデックス追加
CREATE INDEX idx_articles_user_id ON articles(user_id);
CREATE INDEX idx_articles_published_at ON articles(published_at);
```

#### 1-2. SQLクエリの定義

**ファイル**: `db/queries/[feature].sql` （例: `articles.sql`）

```sql
-- name: CreateArticle :one
INSERT INTO articles (user_id, title, content, published_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetArticle :one
SELECT * FROM articles WHERE id = $1 LIMIT 1;

-- name: ListArticles :many
SELECT * FROM articles
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateArticle :one
UPDATE articles
SET title = $2, content = $3, published_at = $4
WHERE id = $1
RETURNING *;

-- name: DeleteArticle :exec
DELETE FROM articles WHERE id = $1;
```

#### 1-3. sqlcコード生成

```bash
make db-generate
```

このコマンドで `internal/db/` 配下に型安全なGoコードが自動生成されます。

---

### 2. Repository層の実装

**ファイル**: `internal/repository/[feature]_repository.go`

#### 役割
- DB操作の抽象化
- sqlc生成コードのラッピング
- インターフェースとして定義し、テスタビリティを確保

#### 実装パターン

```go
package repository

import (
    "context"
    "nanaket-cms/internal/db"
)

// インターフェース定義（依存性の逆転）
type ArticleRepository interface {
    Create(ctx context.Context, userID int64, title, content string, publishedAt *time.Time) (db.Article, error)
    GetByID(ctx context.Context, id int64) (db.Article, error)
    List(ctx context.Context, limit, offset int32) ([]db.Article, error)
    Update(ctx context.Context, id int64, title, content string, publishedAt *time.Time) (db.Article, error)
    Delete(ctx context.Context, id int64) error
}

// 実装構造体
type articleRepository struct {
    querier db.Querier // sqlc生成インターフェース
}

// コンストラクタ
func NewArticleRepository(querier db.Querier) ArticleRepository {
    return &articleRepository{querier: querier}
}

// メソッド実装
func (r *articleRepository) Create(ctx context.Context, userID int64, title, content string, publishedAt *time.Time) (db.Article, error) {
    var nullPublishedAt pgtype.Timestamp
    if publishedAt != nil {
        nullPublishedAt = pgtype.Timestamp{Time: *publishedAt, Valid: true}
    }

    return r.querier.CreateArticle(ctx, db.CreateArticleParams{
        UserID:      userID,
        Title:       title,
        Content:     content,
        PublishedAt: nullPublishedAt,
    })
}

func (r *articleRepository) GetByID(ctx context.Context, id int64) (db.Article, error) {
    return r.querier.GetArticle(ctx, id)
}

// 他のメソッドも同様に実装...
```

---

### 3. Usecase層の実装

**ファイル**: `internal/usecase/[feature]_usecase.go`

#### 役割
- ビジネスロジックの実装
- データ変換・加工
- エラーハンドリング
- 複数のRepositoryを組み合わせた処理

#### 実装パターン

```go
package usecase

import (
    "context"
    "nanaket-cms/internal/db"
    "nanaket-cms/internal/repository"
    "time"
)

// インターフェース定義
type ArticleUsecase interface {
    CreateArticle(ctx context.Context, userID int64, title, content string, publishedAt *time.Time) (db.Article, error)
    GetArticle(ctx context.Context, id int64) (db.Article, error)
    ListArticles(ctx context.Context, limit, offset int32) ([]db.Article, error)
    UpdateArticle(ctx context.Context, id int64, title, content string, publishedAt *time.Time) (db.Article, error)
    DeleteArticle(ctx context.Context, id int64) error
}

// 実装構造体
type articleUsecase struct {
    articleRepo repository.ArticleRepository
    // 必要に応じて他のRepositoryも注入
}

// コンストラクタ
func NewArticleUsecase(articleRepo repository.ArticleRepository) ArticleUsecase {
    return &articleUsecase{
        articleRepo: articleRepo,
    }
}

// メソッド実装
func (u *articleUsecase) CreateArticle(ctx context.Context, userID int64, title, content string, publishedAt *time.Time) (db.Article, error) {
    // ビジネスロジック（例: バリデーション）
    if len(title) == 0 {
        return db.Article{}, errors.New("title is required")
    }
    if len(content) == 0 {
        return db.Article{}, errors.New("content is required")
    }

    // Repositoryに委譲
    return u.articleRepo.Create(ctx, userID, title, content, publishedAt)
}

func (u *articleUsecase) GetArticle(ctx context.Context, id int64) (db.Article, error) {
    return u.articleRepo.GetByID(ctx, id)
}

// 他のメソッドも同様に実装...
```

---

### 4. Handler層の実装

**ファイル**: `internal/handler/[feature]_handler.go`

#### 役割
- HTTPリクエストの解析
- リクエストボディのバリデーション
- Usecaseの呼び出し
- HTTPレスポンスの生成
- ステータスコードの設定

#### 実装パターン

```go
package handler

import (
    "encoding/json"
    "net/http"
    "nanaket-cms/internal/usecase"
    "strconv"
    "time"
)

// リクエスト/レスポンス用の構造体
type CreateArticleRequest struct {
    UserID      int64      `json:"user_id"`
    Title       string     `json:"title"`
    Content     string     `json:"content"`
    PublishedAt *time.Time `json:"published_at,omitempty"`
}

type ArticleResponse struct {
    ID          int64      `json:"id"`
    UserID      int64      `json:"user_id"`
    Title       string     `json:"title"`
    Content     string     `json:"content"`
    PublishedAt *time.Time `json:"published_at,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}

// ハンドラー構造体
type ArticleHandler struct {
    usecase usecase.ArticleUsecase
}

// コンストラクタ
func NewArticleHandler(usecase usecase.ArticleUsecase) *ArticleHandler {
    return &ArticleHandler{usecase: usecase}
}

// POST /api/v1/articles
func (h *ArticleHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
    var req CreateArticleRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    article, err := h.usecase.CreateArticle(r.Context(), req.UserID, req.Title, req.Content, req.PublishedAt)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(ArticleResponse{
        ID:          article.ID,
        UserID:      article.UserID,
        Title:       article.Title,
        Content:     article.Content,
        PublishedAt: convertTimestamp(article.PublishedAt),
        CreatedAt:   article.CreatedAt.Time,
        UpdatedAt:   article.UpdatedAt.Time,
    })
}

// GET /api/v1/articles/{id}
func (h *ArticleHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
    idStr := r.PathValue("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid article ID", http.StatusBadRequest)
        return
    }

    article, err := h.usecase.GetArticle(r.Context(), id)
    if err != nil {
        http.Error(w, "Article not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ArticleResponse{
        ID:          article.ID,
        UserID:      article.UserID,
        Title:       article.Title,
        Content:     article.Content,
        PublishedAt: convertTimestamp(article.PublishedAt),
        CreatedAt:   article.CreatedAt.Time,
        UpdatedAt:   article.UpdatedAt.Time,
    })
}

// ヘルパー関数
func convertTimestamp(ts pgtype.Timestamp) *time.Time {
    if !ts.Valid {
        return nil
    }
    return &ts.Time
}

// 他のメソッドも同様に実装...
```

---

### 5. ルーティング登録

**ファイル**: `cmd/api/main.go`

#### setupRoutes関数に追加

```go
func setupRoutes(mux *http.ServeMux, pool *pgxpool.Pool) {
    // 既存のユーザーAPI
    queries := db.New(pool)
    userRepo := repository.NewUserRepository(queries)
    userUsecase := usecase.NewUserUsecase(userRepo)
    userHandler := handler.NewUserHandler(userUsecase)

    mux.HandleFunc("POST /api/v1/users", userHandler.CreateUser)
    mux.HandleFunc("GET /api/v1/users", userHandler.ListUsers)
    mux.HandleFunc("GET /api/v1/users/{id}", userHandler.GetUser)
    mux.HandleFunc("PUT /api/v1/users/{id}", userHandler.UpdateUser)
    mux.HandleFunc("DELETE /api/v1/users/{id}", userHandler.DeleteUser)

    // 新規追加: 記事API
    articleRepo := repository.NewArticleRepository(queries)
    articleUsecase := usecase.NewArticleUsecase(articleRepo)
    articleHandler := handler.NewArticleHandler(articleUsecase)

    mux.HandleFunc("POST /api/v1/articles", articleHandler.CreateArticle)
    mux.HandleFunc("GET /api/v1/articles", articleHandler.ListArticles)
    mux.HandleFunc("GET /api/v1/articles/{id}", articleHandler.GetArticle)
    mux.HandleFunc("PUT /api/v1/articles/{id}", articleHandler.UpdateArticle)
    mux.HandleFunc("DELETE /api/v1/articles/{id}", articleHandler.DeleteArticle)
}
```

---

## 各層の役割と責務

| 層 | パス | 責務 | 依存先 |
|---|------|------|--------|
| **Handler** | `internal/handler/` | ・HTTPリクエスト/レスポンス処理<br>・JSONエンコード/デコード<br>・バリデーション<br>・ステータスコード設定 | Usecaseインターフェース |
| **Usecase** | `internal/usecase/` | ・ビジネスロジック<br>・データ変換・加工<br>・エラーハンドリング<br>・複数Repositoryの組み合わせ | Repositoryインターフェース |
| **Repository** | `internal/repository/` | ・DB操作の抽象化<br>・sqlcコードのラッピング<br>・クエリ実行 | sqlc生成コード (db.Querier) |
| **DB** | `internal/db/` | ・型安全なクエリ実行<br>・sqlcによる自動生成 | PostgreSQL |

### 依存性の方向

```
Handler → Usecase Interface → Repository Interface → DB (sqlc)
  ↑          ↑                   ↑
実装に依存せず、インターフェースに依存（依存性の逆転原則）
```

---

## 編集するファイルと順序

新しいAPI（例: `articles`）を追加する際の作業順序とファイルリスト：

### ステップ1: データベース設計

| # | ファイル | 作業内容 |
|---|----------|----------|
| 1 | `db/schema/schema.sql` | テーブル定義、インデックス、外部キー制約を追加 |
| 2 | `db/queries/articles.sql` | **新規作成** - CRUD用SQLクエリを定義 |
| 3 | **コマンド実行** | `make db-migrate` でスキーマ適用 |
| 4 | **コマンド実行** | `make db-generate` でGoコード生成 (`internal/db/`) |

### ステップ2: Repository層実装

| # | ファイル | 作業内容 |
|---|----------|----------|
| 5 | `internal/repository/article_repository.go` | **新規作成** - インターフェース定義 + sqlcラッパー実装 |

### ステップ3: Usecase層実装

| # | ファイル | 作業内容 |
|---|----------|----------|
| 6 | `internal/usecase/article_usecase.go` | **新規作成** - ビジネスロジック実装 |

### ステップ4: Handler層実装

| # | ファイル | 作業内容 |
|---|----------|----------|
| 7 | `internal/handler/article_handler.go` | **新規作成** - HTTPハンドラー実装 |

### ステップ5: ルーティング設定

| # | ファイル | 作業内容 |
|---|----------|----------|
| 8 | `cmd/api/main.go` | `setupRoutes()` 関数に新しいエンドポイントを登録 |

### ステップ6: 動作確認

| # | コマンド | 目的 |
|---|----------|------|
| 9 | `make lint` | コード品質チェック |
| 10 | `make run` | アプリケーション起動 |
| 11 | **APIテスト** | curl / Postman等で動作確認 |

---

## 実装例：記事管理API

### データフロー図

```
POST /api/v1/articles
  ↓
[ArticleHandler.CreateArticle]
  ・リクエストボディ解析 (JSON → CreateArticleRequest)
  ・バリデーション（基本チェック）
  ↓
[ArticleUsecase.CreateArticle]
  ・ビジネスロジック（詳細バリデーション）
  ・データ変換
  ↓
[ArticleRepository.Create]
  ・sqlcラッパー呼び出し
  ↓
[db.Querier.CreateArticle]
  ・型安全なSQL実行
  ↓
PostgreSQL
```

### 具体的なコード例

**1. SQL定義** (`db/queries/articles.sql`)

```sql
-- name: CreateArticle :one
INSERT INTO articles (user_id, title, content, published_at)
VALUES ($1, $2, $3, $4)
RETURNING *;
```

**2. Repository** (`internal/repository/article_repository.go`)

```go
func (r *articleRepository) Create(ctx context.Context, userID int64, title, content string, publishedAt *time.Time) (db.Article, error) {
    var nullPublishedAt pgtype.Timestamp
    if publishedAt != nil {
        nullPublishedAt = pgtype.Timestamp{Time: *publishedAt, Valid: true}
    }

    return r.querier.CreateArticle(ctx, db.CreateArticleParams{
        UserID:      userID,
        Title:       title,
        Content:     content,
        PublishedAt: nullPublishedAt,
    })
}
```

**3. Usecase** (`internal/usecase/article_usecase.go`)

```go
func (u *articleUsecase) CreateArticle(ctx context.Context, userID int64, title, content string, publishedAt *time.Time) (db.Article, error) {
    if len(title) == 0 {
        return db.Article{}, errors.New("title is required")
    }
    if len(content) == 0 {
        return db.Article{}, errors.New("content is required")
    }
    return u.articleRepo.Create(ctx, userID, title, content, publishedAt)
}
```

**4. Handler** (`internal/handler/article_handler.go`)

```go
func (h *ArticleHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
    var req CreateArticleRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    article, err := h.usecase.CreateArticle(r.Context(), req.UserID, req.Title, req.Content, req.PublishedAt)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(toArticleResponse(article))
}
```

**5. ルーティング** (`cmd/api/main.go`)

```go
articleRepo := repository.NewArticleRepository(queries)
articleUsecase := usecase.NewArticleUsecase(articleRepo)
articleHandler := handler.NewArticleHandler(articleUsecase)

mux.HandleFunc("POST /api/v1/articles", articleHandler.CreateArticle)
mux.HandleFunc("GET /api/v1/articles/{id}", articleHandler.GetArticle)
```

---

## 命名規約

### ファイル名

| 対象 | 規約 | 例 |
|------|------|-----|
| ファイル名 | `snake_case` | `article_handler.go`, `user_repository.go` |
| パッケージ名 | 小文字単一語 | `handler`, `usecase`, `repository`, `db` |

### Go言語コード

| 対象 | 規約 | 例 |
|------|------|-----|
| インターフェース | `PascalCase` | `ArticleRepository`, `UserUsecase` |
| 構造体 | `PascalCase` | `ArticleHandler`, `CreateUserRequest` |
| 公開関数 | `PascalCase` | `CreateArticle`, `GetUser` |
| 非公開関数 | `camelCase` | `setupRoutes`, `convertTimestamp` |
| 変数 | `camelCase` | `articleRepo`, `userHandler` |

### データベース

| 対象 | 規約 | 例 |
|------|------|-----|
| テーブル名 | `snake_case`（複数形） | `articles`, `users`, `comments` |
| カラム名 | `snake_case` | `user_id`, `created_at`, `published_at` |
| インデックス名 | `idx_[table]_[column]` | `idx_articles_user_id` |

### API URL

| 対象 | 規約 | 例 |
|------|------|-----|
| エンドポイント | `kebab-case`（複数形） | `/api/v1/articles`, `/api/v1/users` |
| リソースID | パラメータ `{id}` | `/api/v1/articles/{id}` |

---

## 開発コマンド

### セットアップ

```bash
# 初回セットアップ（DB起動 + マイグレーション + コード生成）
make dev
```

### データベース操作

```bash
# PostgreSQLコンテナ起動
make db-up

# スキーマ適用（マイグレーション）
make db-migrate

# sqlcコード生成（internal/db/配下に自動生成）
make db-generate

# データベース完全リセット
make db-reset
```

### 開発

```bash
# アプリケーション起動（ポート8080）
make run

# コード検査
make lint

# 自動修正
make lint-fix
```

### テスト

```bash
# 単体テスト実行
make test

# カバレッジ付きテスト
make test-coverage
```

### APIテスト例

```bash
# 記事作成
curl -X POST http://localhost:8080/api/v1/articles \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "title": "サンプル記事",
    "content": "これはテスト記事です",
    "published_at": "2025-11-10T12:00:00Z"
  }'

# 記事取得
curl http://localhost:8080/api/v1/articles/1

# 記事一覧取得
curl http://localhost:8080/api/v1/articles

# 記事更新
curl -X PUT http://localhost:8080/api/v1/articles/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "更新された記事",
    "content": "内容を更新しました"
  }'

# 記事削除
curl -X DELETE http://localhost:8080/api/v1/articles/1
```

---

## チェックリスト

新しいAPI実装時のチェックリストです。

### データベース設計

- [ ] `db/schema/schema.sql` にテーブル定義を追加
- [ ] 必要なインデックスを作成
- [ ] 外部キー制約を設定（必要な場合）
- [ ] `db/queries/[feature].sql` にCRUD用SQLを定義
- [ ] `make db-migrate` でスキーマ適用
- [ ] `make db-generate` でGoコード生成

### Repository層

- [ ] `internal/repository/[feature]_repository.go` を作成
- [ ] インターフェース定義（テスタビリティ確保）
- [ ] sqlcコードをラップした実装
- [ ] コンストラクタ関数の作成
- [ ] エラーハンドリング

### Usecase層

- [ ] `internal/usecase/[feature]_usecase.go` を作成
- [ ] インターフェース定義
- [ ] ビジネスロジック実装
- [ ] バリデーション追加
- [ ] エラーハンドリング

### Handler層

- [ ] `internal/handler/[feature]_handler.go` を作成
- [ ] リクエスト/レスポンス用構造体定義
- [ ] HTTPハンドラー実装
- [ ] 適切なHTTPステータスコード設定
- [ ] エラーレスポンスの実装

### ルーティング

- [ ] `cmd/api/main.go` の `setupRoutes()` を更新
- [ ] 依存性注入の設定
- [ ] エンドポイント登録（POST, GET, PUT, DELETE）

### コード品質

- [ ] `make lint` でコード検査をパス
- [ ] 命名規約に準拠
- [ ] 適切なコメント追加
- [ ] エラーハンドリングの実装

### 動作確認

- [ ] `make run` でアプリケーション起動
- [ ] curl / Postmanでエンドポイントテスト
- [ ] 正常系の動作確認
- [ ] 異常系の動作確認（バリデーションエラー等）
- [ ] レスポンスフォーマットの確認

### ドキュメント

- [ ] API仕様書の更新（必要な場合）
- [ ] README.mdの更新（必要な場合）

---

## 参考資料

### プロジェクト内ドキュメント

- [技術スタック・開発規約](.kiro/steering/tech.md)
- [プロジェクト構造](.kiro/steering/structure.md)
- [プロダクト概要](.kiro/steering/product.md)

### 既存実装例

- **ユーザーAPI**: `internal/handler/user_handler.go`
- **ユーザーUsecase**: `internal/usecase/user_usecase.go`
- **ユーザーRepository**: `internal/repository/user_repository.go`

### 外部ドキュメント

- [sqlc公式ドキュメント](https://docs.sqlc.dev/)
- [pgx（PostgreSQLドライバ）](https://github.com/jackc/pgx)
- [Go標準パッケージ - net/http](https://pkg.go.dev/net/http)

---

## トラブルシューティング

### sqlc生成エラー

```bash
# sqlc.yamlの設定を確認
cat sqlc.yaml

# スキーマファイルの構文チェック
make db-migrate

# 再生成
make db-generate
```

### データベース接続エラー

```bash
# PostgreSQLコンテナの状態確認
docker ps

# コンテナ再起動
make db-reset
make db-up
```

### ルーティングが認識されない

- Go 1.22以降のパターンマッチング機能を使用（`POST /api/v1/users`）
- Goバージョンを確認: `go version`
- パス構文の確認（ワイルドカード `{id}` の位置など）

---

## まとめ

本ガイドに従うことで、一貫性のあるAPI開発が可能になります。

**キーポイント**:
1. **層の責務を明確にする** - Handler / Usecase / Repository / DBの役割分担
2. **インターフェースで抽象化** - テスタビリティと保守性の向上
3. **sqlcで型安全性を確保** - コンパイル時にSQLエラーを検出
4. **命名規約を守る** - コードの可読性向上

質問や改善提案がある場合は、チーム内で議論してこのガイドを継続的に改善していきましょう。
