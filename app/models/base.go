package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base contains common fields for all entities (except ID)
type Base struct {
	HashID    string         `gorm:"type:varchar(20);uniqueIndex" json:"hash_id,omitempty"` // Short public ID for URLs
	IsDummy   bool           `gorm:"default:false" json:"is_dummy,omitempty"`               // Test/seed data flag
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"` // Soft delete
}

// BeforeCreate hook to set HashID if not provided
func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.HashID == "" {
		b.HashID = generateHashID()
	}
	return nil
}

// generateHashID creates a short unique ID (8 chars)
func generateHashID() string {
	id := uuid.New().String()
	return id[:8]
}
