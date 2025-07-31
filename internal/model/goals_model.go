package model

type GoalResponse struct {
	ID        string          `json:"id"`
	MatchID   string          `json:"match_id"`
	PlayerID  string          `json:"player_id"`
	GoalTime  int16           `json:"goal_time"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
	DeletedAt *string         `json:"deleted_at,omitempty"`
	Match     *MatchResponse  `json:"match"`
	Player    *PlayerResponse `json:"player"`
}

type GoalRequestCreate struct {
	MatchID  string `json:"match_id" validate:"required,uuid"`
	PlayerID string `json:"player_id" validate:"required,uuid"`
	GoalTime int16  `json:"goal_time" validate:"required,min=0"`
}

type GoalRequestUpdate struct {
	ID       string `json:"id" validate:"required,uuid"`
	MatchID  string `json:"match_id" validate:"omitempty,uuid"`
	PlayerID string `json:"player_id" validate:"omitempty,uuid"`
	GoalTime int16  `json:"goal_time" validate:"omitempty,min=0"`
}

type GoalRequestFindByID struct {
	ID string `json:"id" validate:"required,uuid"`
}

type GoalRequestSoftDelete struct {
	ID string `json:"id" validate:"required,uuid"`
}
