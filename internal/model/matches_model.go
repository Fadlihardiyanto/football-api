package model

type MatchResponse struct {
	ID         string        `json:"id"`
	MatchDate  string        `json:"match_date"`
	MatchTime  string        `json:"match_time"`
	HomeTeamID string        `json:"home_team_id"`
	AwayTeamID string        `json:"away_team_id"`
	HomeScore  *int          `json:"home_score"`
	AwayScore  *int          `json:"away_score"`
	Status     string        `json:"status"`
	CreatedAt  string        `json:"created_at"`
	UpdatedAt  string        `json:"updated_at"`
	DeletedAt  *string       `json:"deleted_at,omitempty"`
	HomeTeam   *TeamResponse `json:"home_team"`
	AwayTeam   *TeamResponse `json:"away_team"`
}

type MatchRequestCreate struct {
	MatchDate  string `json:"match_date" validate:"required"`
	MatchTime  string `json:"match_time" validate:"required"`
	HomeTeamID string `json:"home_team_id" validate:"required,uuid"`
	AwayTeamID string `json:"away_team_id" validate:"required,uuid"`
	HomeScore  *int   `json:"home_score" validate:"omitempty"`
	AwayScore  *int   `json:"away_score" validate:"omitempty"`
	Status     string `json:"status" validate:"required,oneof=scheduled completed canceled"`
}

type MatchRequestUpdate struct {
	ID         string `json:"id" validate:"required,uuid"`
	MatchDate  string `json:"match_date" validate:"omitempty"`
	MatchTime  string `json:"match_time" validate:"omitempty"`
	HomeTeamID string `json:"home_team_id" validate:"omitempty,uuid"`
	AwayTeamID string `json:"away_team_id" validate:"omitempty,uuid"`
	HomeScore  *int   `json:"home_score" validate:"omitempty"`
	AwayScore  *int   `json:"away_score" validate:"omitempty"`
	Status     string `json:"status" validate:"omitempty,oneof=scheduled completed canceled"`
}

type MatchRequestFindByID struct {
	ID string `json:"id" validate:"required,uuid"`
}

type MatchRequestSoftDelete struct {
	ID string `json:"id" validate:"required,uuid"`
}

type MatchReportResponse struct {
	ID               string       `json:"id"`
	MatchDate        string       `json:"match_date"`
	MatchTime        string       `json:"match_time"`
	HomeTeam         TeamShort    `json:"home_team"`
	AwayTeam         TeamShort    `json:"away_team"`
	HomeScore        int          `json:"home_score"`
	AwayScore        int          `json:"away_score"`
	StatusResult     string       `json:"status_result"`
	Goals            []GoalReport `json:"goals"`
	TopScorer        string       `json:"top_scorer"`
	HomeTeamWinTotal int          `json:"home_team_win_total"`
	AwayTeamWinTotal int          `json:"away_team_win_total"`
}

type TeamShort struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GoalReport struct {
	PlayerName string `json:"player_name"`
	TeamID     string `json:"team_id"`
	Minute     int16  `json:"minute"`
}

type MatchRequestFinish struct {
	ID string `json:"id" validate:"required,uuid"`
}
