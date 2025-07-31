package entity

import (
	"time"
)

type User struct {
	ID           string     `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	Username     string     `gorm:"column:username;unique;not null"`
	Email        string     `gorm:"column:email;unique;not null"`
	PasswordHash string     `gorm:"column:password_hash;not null"`
	Role         string     `gorm:"column:role;not null;default:admin"`
	IsActive     bool       `gorm:"column:is_active;default:true"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
}
