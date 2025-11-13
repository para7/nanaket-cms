-- ユーザー情報テーブル
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,  -- ユーザーID
    name TEXT NOT NULL,                     -- ユーザー名
    email TEXT NOT NULL UNIQUE,             -- メールアドレス
    created_at TEXT NOT NULL DEFAULT (datetime('now')),  -- 作成日時
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))   -- 更新日時
);

-- 記事情報テーブル
CREATE TABLE IF NOT EXISTS articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,  -- 記事ID
    user_id INTEGER NOT NULL REFERENCES users(id),  -- 作成者ID
    title TEXT NOT NULL,                    -- 記事タイトル
    content TEXT NOT NULL,                  -- 記事本文
    published_at TEXT,                      -- 公開日時（NULL = 下書き）
    created_at TEXT NOT NULL DEFAULT (datetime('now')),  -- 作成日時
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))   -- 更新日時
);

-- 作成者による記事検索用インデックス
CREATE INDEX IF NOT EXISTS idx_articles_user_id ON articles(user_id);
-- 公開日時による記事検索用インデックス
CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at);

-- コメント情報テーブル
CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,  -- コメントID
    article_id INTEGER NOT NULL REFERENCES articles(id) ON DELETE CASCADE,  -- 記事ID
    -- 整合性はアプリケーション側で保証
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,  -- コメント作成者ID(ログインしている場合)
    temp_user_name TEXT,                    -- 仮ユーザー名(ログインしていない場合)
    content TEXT NOT NULL,                  -- コメント内容
    created_at TEXT NOT NULL DEFAULT (datetime('now')),  -- 作成日時
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))   -- 更新日時
);

-- 記事によるコメント検索用インデックス
CREATE INDEX IF NOT EXISTS idx_comments_article_id ON comments(article_id);
-- 作成者によるコメント検索用インデックス
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);



-- アクセストークンテーブル
CREATE TABLE IF NOT EXISTS access_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,  -- トークンID
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,  -- ユーザーID
    token TEXT NOT NULL UNIQUE,             -- アクセストークン
    expires_at TEXT,                        -- 有効期限（NULL = 無期限）
    created_at TEXT NOT NULL DEFAULT (datetime('now'))  -- 作成日時
);

-- トークン検索用インデックス
CREATE INDEX IF NOT EXISTS idx_access_tokens_token ON access_tokens(token);
-- ユーザーIDによる検索用インデックス
CREATE INDEX IF NOT EXISTS idx_access_tokens_user_id ON access_tokens(user_id);