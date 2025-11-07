package services

import (
	"errors"
	"time"

	"keepnode/app/models"
	"keepnode/app/queries"
	"keepnode/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserService handles user-related business logic
type UserService struct {
	DB *gorm.DB
}

// NewUserService creates a new UserService instance
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{DB: db}
}

// SignUp creates a new user account
func (s *UserService) SignUp(signUp *models.UserSignUp) (*models.User, error) {
	// Validate sign up data
	validate := utils.NewValidator()
	if err := validate.Struct(signUp); err != nil {
		return nil, errors.New(utils.ValidatorErrorsString(err))
	}

	// Check if user already exists
	existingUser, _ := (&queries.UserQueries{DB: s.DB}).GetUserByEmail(signUp.Email)
	if existingUser.ID != uuid.Nil {
		return nil, errors.New("user with this email already exists")
	}

	// Create new user
	user := &models.User{
		ID:           uuid.New(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Email:        signUp.Email,
		PasswordHash: utils.GeneratePassword(signUp.Password),
		UserStatus:   1, // 0 == blocked, 1 == active
		UserRole:     "user",
	}

	// Validate user data
	if err := validate.Struct(user); err != nil {
		return nil, errors.New(utils.ValidatorErrorsString(err))
	}

	// Save user to database
	if err := (&queries.UserQueries{DB: s.DB}).CreateUser(user); err != nil {
		return nil, err
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return user, nil
}

// SignIn authenticates a user and returns user data with JWT token
func (s *UserService) SignIn(signIn *models.UserSignIn) (*models.User, string, error) {
	// Validate sign in data
	validate := utils.NewValidator()
	if err := validate.Struct(signIn); err != nil {
		return nil, "", errors.New(utils.ValidatorErrorsString(err))
	}

	// Get user by email
	user, err := (&queries.UserQueries{DB: s.DB}).GetUserByEmail(signIn.Email)
	if err != nil {
		return nil, "", errors.New("user with the given email is not found")
	}

	// Check if user is active
	if user.UserStatus != 1 {
		return nil, "", errors.New("user account is blocked")
	}

	// Compare passwords
	if !utils.ComparePasswords(user.PasswordHash, signIn.Password) {
		return nil, "", errors.New("wrong email address or password")
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID.String(), user.Email)
	if err != nil {
		return nil, "", errors.New("failed to generate authentication token")
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return &user, token, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id uuid.UUID) (*models.User, error) {
	user, err := (&queries.UserQueries{DB: s.DB}).GetUserByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	user, err := (&queries.UserQueries{DB: s.DB}).GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return &user, nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(id uuid.UUID, updateData *models.User) error {
	// Check if user exists
	_, err := (&queries.UserQueries{DB: s.DB}).GetUserByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	// Update user
	if err := (&queries.UserQueries{DB: s.DB}).UpdateUser(id, updateData); err != nil {
		return err
	}

	return nil
}

// DeleteUser deletes a user (soft delete)
func (s *UserService) DeleteUser(id uuid.UUID) error {
	// Check if user exists
	_, err := (&queries.UserQueries{DB: s.DB}).GetUserByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	// Delete user
	if err := (&queries.UserQueries{DB: s.DB}).DeleteUser(id); err != nil {
		return err
	}

	return nil
}
