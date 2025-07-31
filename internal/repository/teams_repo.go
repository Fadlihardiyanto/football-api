package repository

import (
	"github.com/Fadlihardiyanto/football-api/internal/entity"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TeamsRepository interface {
	Repository[entity.Team]
	CheckTeamExistsByTeamID(db *gorm.DB, teamID string) (bool, error)
	CheckTeamExistsByName(db *gorm.DB, name string) (bool, error)
}

type teamsRepoImpl struct {
	Repository[entity.Team]
	Log *logrus.Logger
}

func NewTeamsRepo(db *gorm.DB, log *logrus.Logger) TeamsRepository {
	return &teamsRepoImpl{
		Log:        log,
		Repository: NewRepository[entity.Team](db),
	}
}

func (t *teamsRepoImpl) CheckTeamExistsByTeamID(db *gorm.DB, teamID string) (bool, error) {
	var count int64
	if err := db.Model(&entity.Team{}).Where("id = ? AND deleted_at IS NULL", teamID).Count(&count).Error; err != nil {
		t.Log.Errorf("Failed to check if team exists by team ID %s: %v", teamID, err)
		return false, err
	}
	return count > 0, nil
}

func (t *teamsRepoImpl) CheckTeamExistsByName(db *gorm.DB, name string) (bool, error) {
	var count int64
	if err := db.Model(&entity.Team{}).Where("name = ? AND deleted_at IS NULL", name).Count(&count).Error; err != nil {
		t.Log.Errorf("Failed to check if team exists by name %s: %v", name, err)
		return false, err
	}
	return count > 0, nil
}
