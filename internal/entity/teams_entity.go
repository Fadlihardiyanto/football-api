package entity

import (
	"time"
)

type Team struct {
	ID                  string     `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	Name                string     `gorm:"column:name;size:255;not null"`
	Logo                string     `gorm:"column:logo;size:255"`
	FoundedYear         int        `gorm:"column:founded_year;not null"`
	HeadquartersAddress string     `gorm:"column:headquarters_address;type:text;not null"`
	HeadquartersCity    string     `gorm:"column:headquarters_city;size:100;not null"`
	CreatedAt           time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt           *time.Time `gorm:"column:deleted_at"`
	Players             []Player   `gorm:"foreignKey:TeamID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
