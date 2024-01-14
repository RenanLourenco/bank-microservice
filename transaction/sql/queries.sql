-- name: InsertTransaction :exec
INSERT INTO transactions (value,from_user_id,to_user_id) VALUES(?,?,?);

-- name: InsertNotification :exec
INSERT INTO notifications (transaction_id) VALUES(?);

-- name: InsertBalance :exec
INSERT INTO users_balance (balance, user_id) VALUES(?,?);

-- name: UpdateBalance :exec
UPDATE users_balance SET balance = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?;

-- name: FindTransactionById :one
SELECT * FROM transactions WHERE id = ?;

-- name: FindTransactionsBySenderId :many
SELECT * FROM transactions WHERE from_user_id = ?;

-- name: FindTransactionsByReceiverId :many
SELECT * FROM transactions WHERE to_user_id = ?;

-- name: FindTransactionByTimestamps :many
SELECT * FROM transactions AS t WHERE t.created_at BETWEEN ? AND ?;

-- name: FindBalanceByUserId :one
SELECT * FROM users_balance WHERE user_id = ?;
