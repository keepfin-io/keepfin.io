package services

import (
	"context"
	"errors"

	"keepnode/app/models"
	"keepnode/app/queries"
	"keepnode/pkg/utils"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	DB       *gorm.DB
	validate *validator.Validate
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		DB:       db,
		validate: utils.NewValidator(),
	}
}

func (s *UserService) SignUp(signUp *models.UserSignUp) (*models.User, error) {
	if err := s.validate.Struct(signUp); err != nil {
		return nil, errors.New(utils.ValidatorErrorsString(err))
	}

	existingUser, _ := (&queries.UserQueries{DB: s.DB}).GetUserByEmail(context.Background(), signUp.Email)
	if existingUser.ID != uuid.Nil {
		return nil, errors.New("user with this email already exists")
	}

	user := &models.User{
		Email:        signUp.Email,
		PasswordHash: utils.GeneratePassword(signUp.Password),
		UserStatus:   1,
		UserRole:     "user",
		Theme:        models.ThemeLight,
		BaseCurrency: "USD",
		Locale:       models.LocaleEnUS,
	}

	if err := s.validate.Struct(user); err != nil {
		return nil, errors.New(utils.ValidatorErrorsString(err))
	}

	if err := (&queries.UserQueries{DB: s.DB}).CreateUser(context.Background(), user); err != nil {
		return nil, err
	}

	user.PasswordHash = ""

	return user, nil
}

func (s *UserService) SignIn(signIn *models.UserSignIn) (*models.User, string, error) {
	if err := s.validate.Struct(signIn); err != nil {
		return nil, "", errors.New(utils.ValidatorErrorsString(err))
	}

	user, err := (&queries.UserQueries{DB: s.DB}).GetUserByEmail(context.Background(), signIn.Email)
	if err != nil {
		return nil, "", errors.New("user with the given email is not found")
	}

	if user.UserStatus != 1 {
		return nil, "", errors.New("user account is blocked")
	}

	if !utils.ComparePasswords(user.PasswordHash, signIn.Password) {
		return nil, "", errors.New("wrong email address or password")
	}

	token, err := utils.GenerateJWT(user.ID.String(), user.Email)
	if err != nil {
		return nil, "", errors.New("failed to generate authentication token")
	}

	user.PasswordHash = ""

	return user, token, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := (&queries.UserQueries{DB: s.DB}).GetUserByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.PasswordHash = ""

	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := (&queries.UserQueries{DB: s.DB}).GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.PasswordHash = ""

	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, updateData *models.User) error {
	_, err := (&queries.UserQueries{DB: s.DB}).GetUserByID(ctx, id)
	if err != nil {
		return errors.New("user not found")
	}

	// Update user
	if err := (&queries.UserQueries{DB: s.DB}).UpdateUser(ctx, id, updateData); err != nil {
		return err
	}

	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// Check if user exists
	_, err := (&queries.UserQueries{DB: s.DB}).GetUserByID(ctx, id)
	if err != nil {
		return errors.New("user not found")
	}

	// Delete user
	if err := (&queries.UserQueries{DB: s.DB}).DeleteUser(ctx, id); err != nil {
		return err
	}

	return nil
}

func (s *UserService) UpdateProfile(ctx context.Context, id uuid.UUID, req *models.UpdateProfileRequest) (*models.User, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, errors.New(utils.ValidatorErrorsString(err))
	}

	user, err := (&queries.UserQueries{DB: s.DB}).GetUserByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Theme != "" {
		user.Theme = req.Theme
	}
	if req.BaseCurrency != "" {
		user.BaseCurrency = req.BaseCurrency
	}
	if req.Locale != "" {
		user.Locale = req.Locale
	}

	if err := (&queries.UserQueries{DB: s.DB}).UpdateUser(ctx, id, user); err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *UserService) UpdateUserBaseCurrency(ctx context.Context, id uuid.UUID, req *models.UpdateBaseCurrencyRequest) (*models.User, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, errors.New(utils.ValidatorErrorsString(err))
	}

	user, err := (&queries.UserQueries{DB: s.DB}).GetUserByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.BaseCurrency = req.BaseCurrency

	if err := (&queries.UserQueries{DB: s.DB}).UpdateUser(ctx, id, user); err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *UserService) UpdateUserTheme(ctx context.Context, id uuid.UUID, req *models.UpdateThemeRequest) (*models.User, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, errors.New(utils.ValidatorErrorsString(err))
	}

	user, err := (&queries.UserQueries{DB: s.DB}).GetUserByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.Theme = req.Theme

	if err := (&queries.UserQueries{DB: s.DB}).UpdateUser(ctx, id, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUserLocale(ctx context.Context, id uuid.UUID, req *models.UpdateLocaleRequest) (*models.User, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, errors.New(utils.ValidatorErrorsString(err))
	}

	user, err := (&queries.UserQueries{DB: s.DB}).GetUserByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.Locale = req.Locale

	if err := (&queries.UserQueries{DB: s.DB}).UpdateUser(ctx, id, user); err != nil {
		return nil, err
	}

	return user, nil
}
