-- 初期データ投入用SQL
-- このファイルは開発・テスト環境用です

-- テストユーザーの作成
INSERT INTO users (name, email) VALUES
  ('ユーザー1', 'test@example.com')
ON CONFLICT (email) DO NOTHING;

-- テスト用アクセストークンの作成
-- ユーザーIDは users テーブルから取得
INSERT INTO access_tokens (user_id, token, expires_at) VALUES
    (1, 'test-token-1', NULL);