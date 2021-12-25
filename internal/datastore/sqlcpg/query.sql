-- Users

-- name: SaveUser :one
INSERT INTO 
"users" (id, name, email, role) 
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: FindUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: FindUserByEmail :one
SELECT * FROM users WHERE email = $1;


-- Todos

-- name: SaveTodo :one
INSERT INTO todos 
    (id, 	user_id, detail,  done) VALUES
    ( $1,		 $2, 	 $3, 	$4)
RETURNING *;

-- name: FindTodoByID :one
SELECT * FROM todos WHERE id = $1;

-- name: FindAllUserTodos :many
SELECT * FROM todos WHERE 
    user_id = @user_id
ORDER BY id ASC LIMIT @size;

-- name: FindAllUserTodosAsc :many
SELECT * FROM todos WHERE 
    user_id = @user_id
    and id > @cursor
ORDER BY id ASC LIMIT @size;

-- name: FindAllUserTodosDesc :many
SELECT * FROM todos WHERE 
    user_id = @user_id
    and id < @cursor
ORDER BY id ASC LIMIT @size;

-- name: CountAllTodos :one
SELECT COUNT(1) FROM todos WHERE user_id = @user_id;

-- Sessions

-- name: SaveSession :one
INSERT INTO "sessions" 
    (
        id, 
        user_id, 
        refresh_token_expired_at, 
        refresh_token, 
        access_token_expired_at, 
        access_token
    ) VALUES
    (
        @id,
        @user_id,
        @refresh_token_expired_at,
        @refresh_token,
        @access_token_expired_at,
        @access_token
    )
RETURNING *;

-- name: FindSessionByRefreshToken :one 
SELECT * FROM sessions WHERE refresh_token = @refresh_token;

-- name: DeleteSessionByUserID :one
DELETE FROM "sessions" WHERE user_id = @user_id
RETURNING *;

-- name: UpdateAccessToken :one
UPDATE sessions SET access_token = $1, access_token_expired_at = $2 WHERE id = $3
RETURNING *;