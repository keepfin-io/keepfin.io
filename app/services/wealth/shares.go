package wealth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	accountmodel "keepnode/app/models/account"
	usermodel "keepnode/app/models/user"
	model "keepnode/app/models/wealth"
	portfolioservice "keepnode/app/services/portfolio"
)

type ShareRequest struct {
	Name           string      `json:"name"`
	AccountIDs     []uuid.UUID `json:"account_ids"`
	RecipientEmail string      `json:"recipient_email"`
	ExpiresAt      time.Time   `json:"expires_at"`
}

func (s *Service) Shares(user uuid.UUID) ([]model.Share, error) {
	rows := []model.Share{}
	err := s.DB.Where("user_id = ?", user).Order("created_at desc").Find(&rows).Error
	return rows, err
}
func (s *Service) CreateShare(user uuid.UUID, req ShareRequest) (*model.Share, string, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > 100 || len(req.AccountIDs) == 0 || len(req.AccountIDs) > 50 || !req.ExpiresAt.After(time.Now()) || req.ExpiresAt.After(time.Now().AddDate(1, 0, 0)) {
		return nil, "", errors.New("choose accounts, a name and an expiry within one year")
	}
	ids := map[uuid.UUID]bool{}
	for _, id := range req.AccountIDs {
		ids[id] = true
	}
	if len(ids) != len(req.AccountIDs) {
		return nil, "", errors.New("duplicate account")
	}
	var count int64
	if err := s.DB.Model(&accountmodel.Account{}).Where("user_id = ? AND id IN ?", user, req.AccountIDs).Count(&count).Error; err != nil {
		return nil, "", err
	}
	if count != int64(len(req.AccountIDs)) {
		return nil, "", ErrNotFound
	}
	item := model.Share{UserID: user, Name: req.Name, AccountIDs: req.AccountIDs, ExpiresAt: req.ExpiresAt}
	if req.RecipientEmail != "" {
		var recipient usermodel.User
		if err := s.DB.Where("lower(email) = ? AND user_status = 1 AND auth_user_id IS NOT NULL", strings.ToLower(strings.TrimSpace(req.RecipientEmail))).First(&recipient).Error; err != nil {
			return nil, "", errors.New("recipient must have an active Keepfin account")
		}
		item.RecipientID = &recipient.ID
	}
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return nil, "", err
	}
	token := hex.EncodeToString(raw[:])
	hash := sha256.Sum256([]byte(token))
	item.TokenHash = hex.EncodeToString(hash[:])
	if err := s.DB.Create(&item).Error; err != nil {
		return nil, "", err
	}
	return &item, token, nil
}
func (s *Service) RevokeShare(user, id uuid.UUID) error {
	result := s.DB.Model(&model.Share{}).Where("id = ? AND user_id = ?", id, user).Update("revoked_at", time.Now().UTC())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
func (s *Service) SharedReport(token string, viewer *uuid.UUID, period string) (*portfolioservice.PortfolioReport, error) {
	if len(token) != 64 {
		return nil, ErrNotFound
	}
	if _, err := hex.DecodeString(token); err != nil {
		return nil, ErrNotFound
	}
	hash := sha256.Sum256([]byte(token))
	var share model.Share
	if err := s.DB.Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hex.EncodeToString(hash[:]), time.Now().UTC()).First(&share).Error; err != nil {
		return nil, ErrNotFound
	}
	if share.RecipientID != nil && (viewer == nil || (*viewer != *share.RecipientID && *viewer != share.UserID)) {
		return nil, ErrNotFound
	}
	var owner usermodel.User
	if err := s.DB.Where("id = ? AND user_status = 1", share.UserID).First(&owner).Error; err != nil {
		return nil, ErrNotFound
	}
	var count int64
	if err := s.DB.Model(&accountmodel.Account{}).Where("id IN ? AND user_id = ?", share.AccountIDs, share.UserID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 || count != int64(len(share.AccountIDs)) {
		return nil, ErrNotFound
	}
	return s.Portfolio.Report(share.UserID, nil, period, "", portfolioservice.ReportOptions{AccountIDs: share.AccountIDs})
}
