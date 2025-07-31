package converter

import (
	"time"

	"github.com/Fadlihardiyanto/football-api/internal/common"
	"github.com/Fadlihardiyanto/football-api/internal/entity"
	"github.com/Fadlihardiyanto/football-api/internal/model"
)

func ToGoalResponse(goals *entity.Goal) *model.GoalResponse {
	if goals == nil {
		return nil
	}

	return &model.GoalResponse{
		ID:        goals.ID,
		MatchID:   goals.MatchID,
		PlayerID:  goals.PlayerID,
		GoalTime:  goals.GoalTime,
		CreatedAt: goals.CreatedAt.Format(time.RFC3339),
		UpdatedAt: goals.UpdatedAt.Format(time.RFC3339),
		DeletedAt: common.ToStringPointer(goals.DeletedAt),
		Match:     ToMatchResponse(goals.Match),
		Player:    ToPlayerResponse(goals.Player),
	}
}
