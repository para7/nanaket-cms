-- users テーブルの updated_at 自動更新トリガー
DROP TRIGGER IF EXISTS update_users_updated_at;
CREATE TRIGGER update_users_updated_at
AFTER UPDATE ON users
FOR EACH ROW
BEGIN
    UPDATE users SET updated_at = datetime('now')
    WHERE id = NEW.id;
END;

-- articles テーブルの updated_at 自動更新トリガー
DROP TRIGGER IF EXISTS update_articles_updated_at;
CREATE TRIGGER update_articles_updated_at
AFTER UPDATE ON articles
FOR EACH ROW
BEGIN
    UPDATE articles SET updated_at = datetime('now')
    WHERE id = NEW.id;
END;

-- comments テーブルの updated_at 自動更新トリガー
DROP TRIGGER IF EXISTS update_comments_updated_at;
CREATE TRIGGER update_comments_updated_at
AFTER UPDATE ON comments
FOR EACH ROW
BEGIN
    UPDATE comments SET updated_at = datetime('now')
    WHERE id = NEW.id;
END;
