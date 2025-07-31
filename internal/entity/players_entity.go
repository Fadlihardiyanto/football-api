package entity

import (
	"time"
)

type Player struct {
	ID           string     `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	TeamID       string     `gorm:"column:team_id;type:uuid;not null"`
	Name         string     `gorm:"column:name;type:varchar(255);not null"`
	Height       float64    `gorm:"column:height;type:decimal(5,2);not null"`
	Weight       float64    `gorm:"column:weight;type:decimal(5,2);not null"`
	Position     string     `gorm:"column:position;type:varchar(20);not null;check:position IN ('penyerang','gelandang','bertahan','penjaga_gawang')"`
	JerseyNumber int        `gorm:"column:jersey_number;not null"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
	Team         *Team      `gorm:"foreignKey:TeamID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
