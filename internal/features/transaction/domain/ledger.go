package domain

import "time"

type LedgerEntry struct {
	ID            int       `gorm:"primaryKey"`
	UserID        int       `gorm:"not null;index"`
	ReferenceType string    `gorm:"not null"`
	ReferenceID   int       `gorm:"not null;uniqueIndex:idx_ref_type_id"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}
