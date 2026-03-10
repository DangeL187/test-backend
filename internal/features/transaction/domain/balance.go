package domain

type Balance struct {
	UserID   int     `gorm:"primaryKey;column:user_id"`
	Currency string  `gorm:"primaryKey;column:currency"`
	Balance  float64 `gorm:"not null"`
}
