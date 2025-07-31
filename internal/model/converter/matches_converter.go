package converter

import (
	"time"

	"github.com/Fadlihardiyanto/football-api/internal/common"
	"github.com/Fadlihardiyanto/football-api/internal/entity"
	"github.com/Fadlihardiyanto/football-api/internal/model"
)

func ToMatchResponse(match *entity.Match) *model.MatchResponse {
	if match == nil {
		return nil
	}

	return &model.MatchResponse{
		ID:         match.ID,
		HomeTeam:   ToTeamResponse(&match.HomeTeam),
		AwayTeam:   ToTeamResponse(&match.AwayTeam),
		MatchDate:  match.MatchDate.Format("2006-01-02"),
		MatchTime:  match.MatchTime,
		HomeTeamID: match.HomeTeamID,
		AwayTeamID: match.AwayTeamID,
		HomeScore:  match.HomeScore,
		AwayScore:  match.AwayScore,
		Status:     match.Status,
		CreatedAt:  match.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  match.UpdatedAt.Format(time.RFC3339),
		DeletedAt:  common.ToStringPointer(match.DeletedAt),
	}
}

func ToMatchReportResponse(match *entity.Match, goals []entity.Goal, allPreviousMatches []entity.Match) *model.MatchReportResponse {
	var status string
	if *match.HomeScore > *match.AwayScore {
		status = "Home Win"
	} else if *match.HomeScore < *match.AwayScore {
		status = "Away Win"
	} else {
		status = "Draw"
	}

	// count goals per player
	golPerPemain := map[string]int{}
	for _, goal := range goals {
		golPerPemain[goal.PlayerID]++
	}

	// Find top scorer
	var topScorerName string
	var maxGoals int
	for _, goal := range goals {
		count := golPerPemain[goal.PlayerID]
		if count > maxGoals {
			maxGoals = count
			topScorerName = goal.Player.Name
		}
	}

	// Count accumulative wins
	homeWins := 0
	awayWins := 0
	for _, m := range allPreviousMatches {
		if m.HomeScore == nil || m.AwayScore == nil {
			continue
		}
		if *m.HomeScore > *m.AwayScore && m.HomeTeamID == match.HomeTeamID {
			homeWins++
		}
		if *m.AwayScore > *m.HomeScore && m.AwayTeamID == match.AwayTeamID {
			awayWins++
		}
	}

	// Build goal detail
	var goalReports []model.GoalReport
	for _, goal := range goals {
		goalReports = append(goalReports, model.GoalReport{
			PlayerName: goal.Player.Name,
			TeamID:     goal.Player.TeamID,
			Minute:     int16(goal.GoalTime),
		})
	}

	return &model.MatchReportResponse{
		ID:               match.ID,
		MatchDate:        match.MatchDate.Format("2006-01-02"),
		MatchTime:        match.MatchTime,
		HomeTeam:         model.TeamShort{ID: match.HomeTeam.ID, Name: match.HomeTeam.Name},
		AwayTeam:         model.TeamShort{ID: match.AwayTeam.ID, Name: match.AwayTeam.Name},
		HomeScore:        *match.HomeScore,
		AwayScore:        *match.AwayScore,
		StatusResult:     status,
		Goals:            goalReports,
		TopScorer:        topScorerName,
		HomeTeamWinTotal: homeWins,
		AwayTeamWinTotal: awayWins,
	}
}
