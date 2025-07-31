package repository

import (
	"time"

	"github.com/Fadlihardiyanto/football-api/internal/entity"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MatchesRepository interface {
	Repository[entity.Match]
	FindGoalsByMatchIDWithPlayer(tx *gorm.DB, matchID string) ([]entity.Goal, error)
	FindAllBeforeDate(tx *gorm.DB, date time.Time) ([]entity.Match, error)
}

type matchesRepoImpl struct {
	Repository[entity.Match]
	Log *logrus.Logger
}

func NewMatchesRepo(db *gorm.DB, log *logrus.Logger) MatchesRepository {
	return &matchesRepoImpl{
		Log:        log,
		Repository: NewRepository[entity.Match](db),
	}
}

func (r *matchesRepoImpl) FindGoalsByMatchIDWithPlayer(tx *gorm.DB, matchID string) ([]entity.Goal, error) {
	var goals []entity.Goal
	err := tx.Preload("Player").Where("match_id = ?", matchID).Find(&goals).Error
	return goals, err
}

func (r *matchesRepoImpl) FindAllBeforeDate(tx *gorm.DB, date time.Time) ([]entity.Match, error) {
	var matches []entity.Match
	if err := tx.
		Preload("HomeTeam").Preload("AwayTeam").Where("match_date < ?", date).Where("deleted_at IS NULL").Order("match_date ASC").Find(&matches).Error; err != nil {
		return nil, err
	}
	return matches, nil
}
