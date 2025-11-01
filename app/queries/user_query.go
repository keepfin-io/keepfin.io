package queries

import (
	"keepnode/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserQueries struct for queries from User model.
type UserQueries struct {
	DB *gorm.DB
}

// GetUserByID query for getting one User by given ID.
func (q *UserQueries) GetUserByID(id uuid.UUID) (models.User, error) {
	// Define User variable.
	user := models.User{}

	// Send query to database.
	err := q.DB.Where("id = ?", id).First(&user).Error
	if err != nil {
		// Return empty object and error.
		return user, err
	}

	// Return query result.
	return user, nil
}

// GetUserByEmail query for getting one User by given Email.
func (q *UserQueries) GetUserByEmail(email string) (models.User, error) {
	// Define User variable.
	user := models.User{}

	// Send query to database.
	err := q.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		// Return empty object and error.
		return user, err
	}

	// Return query result.
	return user, nil
}

// CreateUser query for creating a new user.
func (q *UserQueries) CreateUser(u *models.User) error {
	// Create user in database.
	err := q.DB.Create(&u).Error
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}

// UpdateUser query for updating user by given ID.
func (q *UserQueries) UpdateUser(id uuid.UUID, u *models.User) error {
	// Update user in database.
	err := q.DB.Model(&models.User{}).Where("id = ?", id).Updates(u).Error
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}

// DeleteUser query for deleting user by given ID (soft delete).
func (q *UserQueries) DeleteUser(id uuid.UUID) error {
	// Soft delete user in database.
	err := q.DB.Delete(&models.User{}, id).Error
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}
