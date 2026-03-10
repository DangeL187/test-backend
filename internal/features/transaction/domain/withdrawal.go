package domain

import "time"

type Withdrawal struct {
	ID             int       `gorm:"primaryKey"`
	UserID         int       `gorm:"not null;index"`
	Amount         float64   `gorm:"not null"`
	Currency       string    `gorm:"not null"`
	Destination    string    `gorm:"not null"`
	Status         string    `gorm:"not null"`
	IdempotencyKey string    `gorm:"not null;uniqueIndex:idx_user_idempotency,user_id"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

type CreateWithdrawalRequest struct {
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	Destination    string  `json:"destination"`
	IdempotencyKey string  `json:"idempotency_key"`
}
