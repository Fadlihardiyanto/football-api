package converter

import (
	"time"

	"github.com/Fadlihardiyanto/football-api/internal/common"
	"github.com/Fadlihardiyanto/football-api/internal/entity"
	"github.com/Fadlihardiyanto/football-api/internal/model"
)

func ToPlayerResponse(player *entity.Player) *model.PlayerResponse {
	if player == nil {
		return nil
	}

	return &model.PlayerResponse{
		ID:           player.ID,
		Name:         player.Name,
		Position:     player.Position,
		Height:       player.Height,
		Weight:       player.Weight,
		JerseyNumber: player.JerseyNumber,
		TeamID:       player.TeamID,
		CreatedAt:    player.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    player.UpdatedAt.Format(time.RFC3339),
		DeletedAt:    common.ToStringPointer(player.DeletedAt),
		Team:         ToTeamResponse(player.Team),
	}
}
