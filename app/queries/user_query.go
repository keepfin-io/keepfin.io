package queries

import (
	"context"
	"errors"
	"fmt"
	"keepnode/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserQueries struct {
	DB *gorm.DB
}

func (q *UserQueries) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User

	err := q.DB.WithContext(ctx).Where(&models.User{ID: id}).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found (id: %s): %w", id, err)
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

func (q *UserQueries) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	err := q.DB.WithContext(ctx).Where(&models.User{Email: email}).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found (email: %s): %w", email, err)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (q *UserQueries) CreateUser(ctx context.Context, user *models.User) error {
	err := q.DB.WithContext(ctx).Create(user).Error
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (q *UserQueries) UpdateUser(ctx context.Context, id uuid.UUID, user *models.User) error {
	err := q.DB.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Updates(user).Error
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (q *UserQueries) DeleteUser(ctx context.Context, id uuid.UUID) error {
	err := q.DB.WithContext(ctx).Delete(&models.User{}, id).Error
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if q.DB.RowsAffected == 0 {
		return fmt.Errorf("user not found (id: %s)", id)
	}

	return nil
}
