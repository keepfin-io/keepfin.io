package user

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	usermodel "keepnode/app/models/user"
	"keepnode/pkg/authsession"
)

var (
	ErrIdentityConflict = errors.New("existing account requires explicit identity mapping")
	ErrUserBlocked      = errors.New("account unavailable")
)

// ResolveIdentity preserves application UUIDs and never links legacy financial
// records just because a newly authenticated identity has the same email.
func (s *UserService) ResolveIdentity(ctx context.Context, identity *authsession.Identity) (*usermodel.User, error) {
	if identity == nil || identity.ID == "" || !identity.EmailVerified || identity.Email == "" {
		return nil, authsession.ErrUnauthorized
	}
	// Normal requests need only one application-user lookup. Lock only during
	// first-login provisioning, not every parallel dashboard request.
	existing, err := findMappedUser(s.DB.WithContext(ctx), identity.ID)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return existing, err
	}
	var user usermodel.User
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Serialize concurrent first requests (including case variants of an email).
		email := strings.ToLower(identity.Email)
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", email).Error; err != nil {
			return err
		}
		existing, err := findMappedUser(tx, identity.ID)
		if err == nil {
			user = *existing
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var count int64
		if err := tx.Unscoped().Model(&usermodel.User{}).Where("lower(email) = ?", email).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrIdentityConflict
		}
		user = usermodel.User{AuthUserID: &identity.ID, Email: email, UserStatus: 1, UserRole: "user",
			Theme: usermodel.ThemeLight, BaseCurrency: "USD", Locale: usermodel.LocaleEnUS}
		// A unique identity index is the final guard against duplicate provisioning.
		return tx.Create(&user).Error
	})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func findMappedUser(db *gorm.DB, authUserID string) (*usermodel.User, error) {
	var user usermodel.User
	if err := db.Unscoped().Where("auth_user_id = ?", authUserID).First(&user).Error; err != nil {
		return nil, err
	}
	if user.DeletedAt.Valid || user.UserStatus != 1 {
		return nil, ErrUserBlocked
	}
	return &user, nil
}
