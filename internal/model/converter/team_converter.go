package converter

import (
	"time"

	"github.com/Fadlihardiyanto/football-api/internal/common"
	"github.com/Fadlihardiyanto/football-api/internal/entity"
	"github.com/Fadlihardiyanto/football-api/internal/model"
)

func ToTeamResponse(team *entity.Team) *model.TeamResponse {
	if team == nil {
		return nil
	}

	var players []*model.PlayerResponse
	if team.Players != nil {
		players = make([]*model.PlayerResponse, len(team.Players))
		for i, player := range team.Players {
			players[i] = ToPlayerResponse(&player)
		}
	}

	return &model.TeamResponse{
		ID:                  team.ID,
		Name:                team.Name,
		Logo:                team.Logo,
		FoundedYear:         team.FoundedYear,
		HeadquartersAddress: team.HeadquartersAddress,
		HeadquartersCity:    team.HeadquartersCity,
		CreatedAt:           team.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           team.UpdatedAt.Format(time.RFC3339),
		DeletedAt:           common.ToStringPointer(team.DeletedAt),
		Players:             players,
	}
}
