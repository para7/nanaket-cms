-- ユーザー情報テーブル
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,              -- ユーザーID
    name VARCHAR(255) NOT NULL,            -- ユーザー名
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
