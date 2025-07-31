package repository

import (
	"github.com/Fadlihardiyanto/football-api/internal/entity"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UsersRepository interface {
	Repository[entity.User]
	FindByEmail(db *gorm.DB, email string) (*entity.User, error)
}

type userRepoImpl struct {
	Repository[entity.User]
	Log *logrus.Logger
}

func NewUserRepo(db *gorm.DB, log *logrus.Logger) UsersRepository {
	return &userRepoImpl{
		Log:        log,
		Repository: NewRepository[entity.User](db),
	}
}

func (r *userRepoImpl) FindByEmail(db *gorm.DB, email string) (*entity.User, error) {
	var user entity.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		r.Log.Errorf("Failed to find user by email %s: %v", email, err)
		return nil, err
	}
	return &user, nil
}
