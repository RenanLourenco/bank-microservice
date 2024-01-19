-- name: CreateNaturalUser :exec
INSERT INTO users (name, email, password, cpf, user_type) VALUES(?,?,?,?,?);

-- name: CreateLegalUser :exec
INSERT INTO users (name, email, password, cnpj, user_type) VALUES(?,?,?,?,?);

-- name: FindUserById :one
SELECT * FROM users WHERE id = ?;

-- name: FindUserByEmail :one
SELECT * FROM users WHERE email = ?;

-- name: FindNaturalUser :one
SELECT * FROM users WHERE cpf = ?;

-- name: FindLegalUser :one
SELECT * FROM users WHERE cnpj = ?;

-- name: UpdateUser :exec
UPDATE users SET name = COALESCE(?, name), updated_at = NOW() WHERE id = ?;
