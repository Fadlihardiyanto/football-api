package repository

import (
	"github.com/Fadlihardiyanto/football-api/internal/entity"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type GoalsRepository interface {
	Repository[entity.Goal]
}

type goalsRepoImpl struct {
	Repository[entity.Goal]
	Log *logrus.Logger
}

func NewGoalsRepo(db *gorm.DB, log *logrus.Logger) GoalsRepository {
	return &goalsRepoImpl{
		Log:        log,
		Repository: NewRepository[entity.Goal](db),
	}
}
