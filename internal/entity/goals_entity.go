package entity

import (
	"time"
)

type Goal struct {
	ID        string     `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	MatchID   string     `gorm:"column:match_id;type:uuid;not null"`
	PlayerID  string     `gorm:"column:player_id;type:uuid;not null"`
	GoalTime  int16      `gorm:"column:goal_time;type:smallint;not null"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	Match     *Match     `gorm:"foreignKey:MatchID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Player    *Player    `gorm:"foreignKey:PlayerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
