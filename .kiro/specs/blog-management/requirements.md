# Requirements Document

## Project Description (Input)
ブログ管理用のテーブルを一式用意する

## Introduction
Nanaket CMSにブログコンテンツ管理機能を追加します。本仕様では、ブログ記事、カテゴリー、タグ、コメントなど、ブログ運用に必要な基本要素を管理するためのデータベーステーブル定義と、それらを操作するREST APIを実装します。既存のユーザー管理機能と連携し、記事の著者管理も行います。

## Requirements

### Requirement 1: ブログ記事管理
**Objective:** As a コンテンツ編集者, I want ブログ記事を作成・編集・削除・公開管理できる機能, so that CMSを通じてブログコンテンツを効率的に管理できる

#### Acceptance Criteria
1. The Blog Management System shall provide a `posts` table with fields: id, title, slug, content, excerpt, author_id (foreign key to users), status (draft/published/archived), published_at, created_at, updated_at
2. When a new post is created via POST `/api/v1/posts`, the Blog Management System shall store the post with status "draft" by default
3. When a post is updated via PUT `/api/v1/posts/{id}`, the Blog Management System shall update the `updated_at` timestamp
4. When a post status changes to "published", the Blog Management System shall set `published_at` to the current timestamp if not already set
5. When a post is retrieved via GET `/api/v1/posts/{id}`, the Blog Management System shall return the post data including author information
6. When posts are listed via GET `/api/v1/posts`, the Blog Management System shall support filtering by status, author_id, and pagination
7. When a post is deleted via DELETE `/api/v1/posts/{id}`, the Blog Management System shall permanently remove the post and all related data
8. The Blog Management System shall enforce unique `slug` values per post to enable SEO-friendly URLs

### Requirement 2: カテゴリー管理
**Objective:** As a コンテンツ編集者, I want ブログ記事をカテゴリーで分類できる, so that コンテンツを体系的に整理し、読者が関連記事を見つけやすくする

#### Acceptance Criteria
1. The Blog Management System shall provide a `categories` table with fields: id, name, slug, description, parent_id (self-referencing for hierarchy), created_at, updated_at
2. When a new category is created via POST `/api/v1/categories`, the Blog Management System shall validate that the slug is unique
3. When a category with `parent_id` is created, the Blog Management System shall verify the parent category exists
4. The Blog Management System shall provide a `post_categories` junction table to support many-to-many relationship between posts and categories
5. When categories are listed via GET `/api/v1/categories`, the Blog Management System shall return hierarchical structure if `parent_id` relationships exist
6. When a post is associated with categories via POST `/api/v1/posts/{id}/categories`, the Blog Management System shall create entries in `post_categories` table
7. If a category is deleted and has associated posts, then the Blog Management System shall remove all `post_categories` entries for that category

### Requirement 3: タグ管理
**Objective:** As a コンテンツ編集者, I want ブログ記事にタグを付与できる, so that 横断的なトピックで記事を分類し、検索性を向上させる

#### Acceptance Criteria
1. The Blog Management System shall provide a `tags` table with fields: id, name, slug, created_at, updated_at
2. The Blog Management System shall provide a `post_tags` junction table to support many-to-many relationship between posts and tags
3. When a new tag is created via POST `/api/v1/tags`, the Blog Management System shall enforce unique slug values
4. When tags are assigned to a post via POST `/api/v1/posts/{id}/tags`, the Blog Management System shall create or reference existing tags
5. When a post is retrieved, the Blog Management System shall include all associated tags in the response
6. When tags are listed via GET `/api/v1/tags`, the Blog Management System shall support sorting by usage count (number of posts)
7. If a tag is deleted and has no associated posts, then the Blog Management System shall allow deletion; otherwise shall return an error

### Requirement 4: コメント管理
**Objective:** As a ブログ読者, I want 記事にコメントを投稿できる, so that 著者や他の読者と交流できる

#### Acceptance Criteria
1. The Blog Management System shall provide a `comments` table with fields: id, post_id (foreign key to posts), author_name, author_email, content, status (pending/approved/spam), parent_id (self-referencing for replies), created_at, updated_at
2. When a new comment is submitted via POST `/api/v1/posts/{id}/comments`, the Blog Management System shall store it with status "pending"
3. When comments are retrieved for a post via GET `/api/v1/posts/{id}/comments`, the Blog Management System shall return only comments with status "approved"
4. When a comment has `parent_id`, the Blog Management System shall verify the parent comment exists and belongs to the same post
5. If a comment is marked as "spam", then the Blog Management System shall exclude it from all public API responses
6. When a comment is deleted via DELETE `/api/v1/comments/{id}`, the Blog Management System shall also delete all child replies (cascade delete)
7. The Blog Management System shall support threaded comments up to 3 levels deep

### Requirement 5: メディア管理（メタデータ）
**Objective:** As a コンテンツ編集者, I want 記事に使用する画像や動画のメタデータを管理できる, so that メディアファイルとコンテンツの関連付けを明確にする

#### Acceptance Criteria
1. The Blog Management System shall provide a `media` table with fields: id, file_name, file_path, file_type, file_size, alt_text, caption, uploaded_by (foreign key to users), created_at
2. When a media record is created via POST `/api/v1/media`, the Blog Management System shall store metadata (actual file upload is out of scope)
3. The Blog Management System shall provide a `post_media` junction table to associate media with posts
4. When media is associated with a post via POST `/api/v1/posts/{id}/media`, the Blog Management System shall validate that both post and media exist
5. When a post is retrieved, the Blog Management System shall include all associated media metadata
6. When media is deleted via DELETE `/api/v1/media/{id}`, the Blog Management System shall remove all `post_media` associations

### Requirement 6: データ整合性と制約
**Objective:** As a システム管理者, I want データベースレベルで整合性が保証される, so that データ破損や不正なデータ状態を防止できる

#### Acceptance Criteria
1. The Blog Management System shall enforce foreign key constraints for all relationships (author_id, post_id, category_id, tag_id, etc.)
2. When a user is deleted from the `users` table, the Blog Management System shall handle posts by that author according to a defined policy (cascade delete, set null, or prevent deletion)
3. The Blog Management System shall enforce NOT NULL constraints on critical fields: posts.title, posts.slug, posts.author_id, categories.name, tags.name
4. The Blog Management System shall create indexes on frequently queried fields: posts.slug, posts.status, posts.published_at, categories.slug, tags.slug
5. When database operations fail due to constraint violations, the Blog Management System shall return appropriate HTTP error codes (409 Conflict for unique violations, 404 Not Found for foreign key violations)

### Requirement 7: タイムスタンプとソフトデリート（オプション）
**Objective:** As a システム管理者, I want 削除された記事を復元できる可能性を保持する, so that 誤削除からの回復が可能になる

#### Acceptance Criteria
1. Where soft delete is enabled, the Blog Management System shall add `deleted_at` field to posts, categories, tags, and comments tables
2. Where soft delete is enabled, when a record is deleted via DELETE endpoints, the Blog Management System shall set `deleted_at` to current timestamp instead of physically removing the record
3. Where soft delete is enabled, the Blog Management System shall exclude soft-deleted records from all standard queries unless explicitly requested
4. The Blog Management System shall automatically set `created_at` to current timestamp on record creation
5. The Blog Management System shall automatically update `updated_at` to current timestamp on record modification
