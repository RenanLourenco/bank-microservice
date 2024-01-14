
CREATE TABLE IF NOT EXISTS transactions (
  id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  value DECIMAL(10,2) NOT NULL,
  from_user_id INT NOT NULL,
  FOREIGN KEY (from_user_id) REFERENCES users_balance(user_id),
  to_user_id INT NOT NULL,
  FOREIGN KEY (to_user_id) REFERENCES users_balance(user_id),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users_balance (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  balance DECIMAL(10,2),
  user_id INT NOT NULL UNIQUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE notifications (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    transaction_id INT NOT NULL,
    FOREIGN KEY (transaction_id) REFERENCES transactions(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
ENGINE = InnoDB;