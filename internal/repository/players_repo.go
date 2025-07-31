package repository

import (
	"github.com/Fadlihardiyanto/football-api/internal/entity"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PlayersRepository interface {
	Repository[entity.Player]
	FindPlayerWithTeamByID(db *gorm.DB, id string) (*entity.Player, error)
	CheckNumberJerseyByNumberAndTeamID(db *gorm.DB, number int, teamID string) (bool, error)
	CheckPlayerAlreadyHasTeam(db *gorm.DB, playerID string) (bool, error)
	CheckNumberJerseyByNumberAndTeamIDExceptPlayerID(db *gorm.DB, number int, teamID, playerID string) (bool, error)
}

type playersRepoImpl struct {
	Repository[entity.Player]
	Log *logrus.Logger
}

func NewPlayersRepo(db *gorm.DB, log *logrus.Logger) PlayersRepository {
	return &playersRepoImpl{
		Log:        log,
		Repository: NewRepository[entity.Player](db),
	}
}

func (p *playersRepoImpl) FindPlayerWithTeamByID(db *gorm.DB, id string) (*entity.Player, error) {
	var player entity.Player
	if err := db.Preload("Team").Where("id = ? AND deleted_at IS NULL", id).First(&player).Error; err != nil {
		p.Log.Errorf("Failed to find player with team by ID %s: %v", id, err)
		return nil, err
	}
	return &player, nil
}

func (p *playersRepoImpl) CheckNumberJerseyByNumberAndTeamID(db *gorm.DB, number int, teamID string) (bool, error) {
	var count int64
	if err := db.Model(&entity.Player{}).Where("jersey_number = ? AND team_id = ? AND deleted_at IS NULL", number, teamID).Count(&count).Error; err != nil {
		p.Log.Errorf("Failed to check jersey number by number %d and team ID %s: %v", number, teamID, err)
		return false, err
	}
	return count > 0, nil
}

func (p *playersRepoImpl) CheckPlayerAlreadyHasTeam(db *gorm.DB, playerID string) (bool, error) {
	var count int64
	if err := db.Model(&entity.Player{}).Where("id = ? AND team_id IS NOT NULL AND deleted_at IS NULL", playerID).Count(&count).Error; err != nil {
		p.Log.Errorf("Failed to check if player already has a team by player ID %s: %v", playerID, err)
		return false, err
	}
	return count > 0, nil
}

func (p *playersRepoImpl) CheckNumberJerseyByNumberAndTeamIDExceptPlayerID(db *gorm.DB, number int, teamID, playerID string) (bool, error) {
	var count int64
	if err := db.Model(&entity.Player{}).Where("jersey_number = ? AND team_id = ? AND id != ? AND deleted_at IS NULL", number, teamID, playerID).Count(&count).Error; err != nil {
		p.Log.Errorf("Failed to check jersey number by number %d and team ID %s except player ID %s: %v", number, teamID, playerID, err)
		return false, err
	}
	return count > 0, nil
}
