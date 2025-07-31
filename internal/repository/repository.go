package repository

import "gorm.io/gorm"

type Repository[T any] interface {
	FindByID(db *gorm.DB, id string) (*T, error)
	FindByIDWithRelations(db *gorm.DB, id string, relations ...string) (*T, error)
	FindAll(db *gorm.DB) ([]T, error)
	FindAllWithRelations(db *gorm.DB, relations ...string) ([]T, error)
	Update(db *gorm.DB, entity *T) error
	Create(db *gorm.DB, entity *T) error
	Delete(db *gorm.DB, entity *T) error
	SoftDelete(db *gorm.DB, id string) error
}

type repositoryImpl[T any] struct {
	DB *gorm.DB
}

func NewRepository[T any](db *gorm.DB) Repository[T] {
	return &repositoryImpl[T]{DB: db}
}

func (r *repositoryImpl[T]) Create(db *gorm.DB, entity *T) error {
	return db.Create(entity).Error
}

func (r *repositoryImpl[T]) FindByID(db *gorm.DB, id string) (*T, error) {
	var entity T
	if err := db.Where("id = ? and deleted_at is null", id).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *repositoryImpl[T]) FindByIDWithRelations(db *gorm.DB, id string, relations ...string) (*T, error) {
	var entity T
	query := db.Where("id = ? and deleted_at is null", id)
	for _, relation := range relations {
		query = query.Preload(relation)
	}
	if err := query.First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *repositoryImpl[T]) FindAll(db *gorm.DB) ([]T, error) {
	var entities []T
	if err := db.Find(&entities, "deleted_at is null").Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *repositoryImpl[T]) FindAllWithRelations(db *gorm.DB, relations ...string) ([]T, error) {
	var entities []T
	query := db.Model(&entities).Where("deleted_at is null")
	for _, relation := range relations {
		query = query.Preload(relation)
	}
	if err := query.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *repositoryImpl[T]) Update(db *gorm.DB, entity *T) error {
	return db.Save(entity).Error
}

func (r *repositoryImpl[T]) Delete(db *gorm.DB, entity *T) error {
	return db.Delete(entity).Error
}

func (r *repositoryImpl[T]) SoftDelete(db *gorm.DB, id string) error {
	var entity T
	if err := db.Where("id = ?", id).First(&entity).Error; err != nil {
		return err
	}

	if err := db.Model(&entity).Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error; err != nil {
		return err
	}

	return nil
}
