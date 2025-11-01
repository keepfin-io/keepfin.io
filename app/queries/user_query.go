package queries

import (
	"database/sql"

	"keepnode/app/models"

	"github.com/google/uuid"
)

// UserQueries struct for queries from User model.
type UserQueries struct {
	DB *sql.DB
}

// GetUserByID query for getting one User by given ID.
func (q *UserQueries) GetUserByID(id uuid.UUID) (models.User, error) {
	// Define User variable.
	user := models.User{}

	// Define query string.
	query := `SELECT * FROM users WHERE id = $1`

	// Send query to database.
	err := q.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Email,
		&user.PasswordHash,
		&user.UserStatus,
		&user.UserRole,
	)
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

	// Define query string.
	query := `SELECT * FROM users WHERE email = $1`

	// Send query to database.
	err := q.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Email,
		&user.PasswordHash,
		&user.UserStatus,
		&user.UserRole,
	)
	if err != nil {
		// Return empty object and error.
		return user, err
	}

	// Return query result.
	return user, nil
}

// CreateUser query for creating a new user.
func (q *UserQueries) CreateUser(u *models.User) error {
	// Define query string.
	query := `INSERT INTO users (id, created_at, updated_at, email, password_hash, user_status, user_role) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	// Send query to database.
	_, err := q.DB.Exec(
		query,
		u.ID,
		u.CreatedAt,
		u.UpdatedAt,
		u.Email,
		u.PasswordHash,
		u.UserStatus,
		u.UserRole,
	)
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}

// UpdateUser query for updating user by given ID.
func (q *UserQueries) UpdateUser(id uuid.UUID, u *models.User) error {
	// Define query string.
	query := `UPDATE users SET updated_at = $2, email = $3, password_hash = $4, user_status = $5, user_role = $6 WHERE id = $1`

	// Send query to database.
	_, err := q.DB.Exec(query, id, u.UpdatedAt, u.Email, u.PasswordHash, u.UserStatus, u.UserRole)
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}

// DeleteUser query for deleting user by given ID.
func (q *UserQueries) DeleteUser(id uuid.UUID) error {
	// Define query string.
	query := `DELETE FROM users WHERE id = $1`

	// Send query to database.
	_, err := q.DB.Exec(query, id)
	if err != nil {
		// Return only error.
		return err
	}

	// This query returns nothing.
	return nil
}
