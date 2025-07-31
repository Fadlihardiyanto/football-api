package entity

import (
	"time"
)

type Match struct {
	ID         string     `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	MatchDate  time.Time  `gorm:"column:match_date;type:date;not null"`
	MatchTime  string     `gorm:"column:match_time;type:time;not null"`
	HomeTeamID string     `gorm:"column:home_team_id;type:uuid;not null"`
	AwayTeamID string     `gorm:"column:away_team_id;type:uuid;not null"`
	HomeScore  *int       `gorm:"column:home_score"`
	AwayScore  *int       `gorm:"column:away_score"`
	Status     string     `gorm:"column:status;type:varchar(20);default:scheduled"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt  *time.Time `gorm:"column:deleted_at"`
	HomeTeam   Team       `gorm:"foreignKey:HomeTeamID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	AwayTeam   Team       `gorm:"foreignKey:AwayTeamID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
