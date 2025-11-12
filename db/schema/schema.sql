-- ユーザー情報テーブル
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,              -- ユーザーID
    name TEXT NOT NULL,            -- ユーザー名
    email VARCHAR(255) NOT NULL UNIQUE,     -- メールアドレス
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,  -- 作成日時
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP   -- 更新日時
);

-- 記事情報テーブル
CREATE TABLE IF NOT EXISTS articles (
    id BIGSERIAL PRIMARY KEY,              -- 記事ID
    user_id BIGINT NOT NULL REFERENCES users(id),  -- 作成者ID
    title VARCHAR(500) NOT NULL,           -- 記事タイトル
    content TEXT NOT NULL,                 -- 記事本文
    published_at TIMESTAMP,                -- 公開日時（NULL = 下書き）
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,  -- 作成日時
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP   -- 更新日時
);

-- 作成者による記事検索用インデックス
CREATE INDEX IF NOT EXISTS idx_articles_user_id ON articles(user_id);
-- 公開日時による記事検索用インデックス
CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at);

-- コメント情報テーブル
CREATE TABLE IF NOT EXISTS comments (
    id BIGSERIAL PRIMARY KEY,              -- コメントID
    article_id BIGINT NOT NULL REFERENCES articles(id) ON DELETE CASCADE,  -- 記事ID
    -- 整合性はアプリケーション側で保証
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,        -- コメント作成者ID(ログインしている場合)
    temp_user_name VARCHAR(255),          -- 仮ユーザー名(ログインしていない場合) 
    content TEXT NOT NULL,                 -- コメント内容
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,  -- 作成日時
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP   -- 更新日時
);

-- 記事によるコメント検索用インデックス
CREATE INDEX IF NOT EXISTS idx_comments_article_id ON comments(article_id);
-- 作成者によるコメント検索用インデックス
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);



-- アクセストークンテーブル
CREATE TABLE IF NOT EXISTS access_tokens (
    id BIGSERIAL PRIMARY KEY,              -- トークンID
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,  -- ユーザーID
    token VARCHAR(255) NOT NULL UNIQUE,    -- アクセストークン
    expires_at TIMESTAMP,                  -- 有効期限（NULL = 無期限）
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 作成日時
);

-- トークン検索用インデックス
CREATE INDEX IF NOT EXISTS idx_access_tokens_token ON access_tokens(token);
-- ユーザーIDによる検索用インデックス
CREATE INDEX IF NOT EXISTS idx_access_tokens_user_id ON access_tokens(user_id);


-- 記事の下書き・自動保存テーブル
CREATE TABLE IF NOT EXISTS article_drafts (
    id BIGSERIAL PRIMARY KEY,
    article_id BIGINT REFERENCES articles(id) ON DELETE CASCADE,  -- NULL = 新規記事の下書き
    user_id BIGINT NOT NULL REFERENCES users(id),  -- 編集者
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    is_auto_saved BOOLEAN NOT NULL DEFAULT true,  -- true: 自動保存, false: 手動保存
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 保存日時
);

CREATE INDEX IF NOT EXISTS idx_article_drafts_article_id ON article_drafts(article_id);
CREATE INDEX IF NOT EXISTS idx_article_drafts_user_id ON article_drafts(user_id);
CREATE INDEX IF NOT EXISTS idx_article_drafts_created_at ON article_drafts(created_at);

-- 公開記事の履歴テーブル
CREATE TABLE IF NOT EXISTS article_histories (
    id BIGSERIAL PRIMARY KEY,  -- IDの時系列順でバージョン管理
    article_id BIGINT NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id),  -- 更新者
    is_auto_saved BOOLEAN NOT NULL DEFAULT false,  -- 自動更新か手動更新か
    published_at TIMESTAMP NOT NULL,  -- 公開日時
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 履歴作成日時
);

CREATE INDEX IF NOT EXISTS idx_article_histories_article_id ON article_histories(article_id);
CREATE INDEX IF NOT EXISTS idx_article_histories_created_at ON article_histories(created_at);