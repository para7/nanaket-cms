-- name: GetArticle :one
SELECT * FROM articles
WHERE id = $1 LIMIT 1;

-- name: ListArticles :many
SELECT * FROM articles
ORDER BY id;

-- name: CreateArticle :one
INSERT INTO articles (
    user_id, title, content, published_at
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateArticle :one
UPDATE articles
SET user_id = $1, title = $2, content = $3, published_at = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $5
RETURNING *;

-- name: DeleteArticle :exec
DELETE FROM articles
WHERE id = $1;

-- name: ListArticlesByUser :many
SELECT * FROM articles
WHERE user_id = $1
ORDER BY id;

