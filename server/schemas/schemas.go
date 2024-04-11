package schemas

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Token   string `json:"token"`
	IsAdmin bool   `json:"is_admin"`
}

type Banner struct {
	ID        uint           `gorm:"primaryKey" json:"banner_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	FeatureID int            `json:"feature_id"`
	IsActive  bool           `json:"is_active"`
	TagIDs    pq.Int64Array  `gorm:"type:integer []" json:"tag_ids"`
	Content   JSONB          `gorm:"type:jsonb" json:"content"`
}
