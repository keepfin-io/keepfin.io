package user

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	usermodel "keepnode/app/models/user"
	userquery "keepnode/app/queries/user"
	"keepnode/pkg/utils"
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

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*usermodel.User, error) {
	user, err := (&userquery.UserQueries{DB: s.DB}).GetUserByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*usermodel.User, error) {
	user, err := (&userquery.UserQueries{DB: s.DB}).GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, updateData *usermodel.User) error {
	_, err := (&userquery.UserQueries{DB: s.DB}).GetUserByID(ctx, id)
	if err != nil {
		return errors.New("user not found")
	}

	// Update user
	if err := (&userquery.UserQueries{DB: s.DB}).UpdateUser(ctx, id, updateData); err != nil {
		return err
	}

	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// Check if user exists
	_, err := (&userquery.UserQueries{DB: s.DB}).GetUserByID(ctx, id)
	if err != nil {
		return errors.New("user not found")
	}

	// Delete user
	if err := (&userquery.UserQueries{DB: s.DB}).DeleteUser(ctx, id); err != nil {
		return err
	}

	return nil
}

func (s *UserService) UpdateProfile(ctx context.Context, id uuid.UUID, req *usermodel.UpdateProfileRequest) (*usermodel.User, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, errors.New(utils.ValidatorErrorsString(err))
	}

	user, err := (&userquery.UserQueries{DB: s.DB}).GetUserByID(ctx, id)
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

	if err := (&userquery.UserQueries{DB: s.DB}).UpdateUser(ctx, id, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUserBaseCurrency(ctx context.Context, id uuid.UUID, req *usermodel.UpdateBaseCurrencyRequest) (*usermodel.User, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, errors.New(utils.ValidatorErrorsString(err))
	}

	user, err := (&userquery.UserQueries{DB: s.DB}).GetUserByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.BaseCurrency = req.BaseCurrency

	if err := (&userquery.UserQueries{DB: s.DB}).UpdateUser(ctx, id, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUserTheme(ctx context.Context, id uuid.UUID, req *usermodel.UpdateThemeRequest) (*usermodel.User, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, errors.New(utils.ValidatorErrorsString(err))
	}

	user, err := (&userquery.UserQueries{DB: s.DB}).GetUserByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.Theme = req.Theme

	if err := (&userquery.UserQueries{DB: s.DB}).UpdateUser(ctx, id, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUserLocale(ctx context.Context, id uuid.UUID, req *usermodel.UpdateLocaleRequest) (*usermodel.User, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, errors.New(utils.ValidatorErrorsString(err))
	}

	user, err := (&userquery.UserQueries{DB: s.DB}).GetUserByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.Locale = req.Locale

	if err := (&userquery.UserQueries{DB: s.DB}).UpdateUser(ctx, id, user); err != nil {
		return nil, err
	}

	return user, nil
}
