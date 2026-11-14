package wealth

import (
	"errors"
	accountmodel "keepnode/app/models/account"
	activitymodel "keepnode/app/models/activity"
	model "keepnode/app/models/wealth"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Service) Tags(user uuid.UUID) ([]model.Tag, error) {
	rows := []model.Tag{}
	err := s.DB.Where("user_id = ?", user).Order("name,id").Find(&rows).Error
	return rows, err
}

// TagColors are the palette keys a tag may use; the client owns the styles.
var TagColors = []string{"gray", "red", "orange", "amber", "green", "teal", "blue", "violet", "pink"}

func validTagColor(color string) bool {
	for _, c := range TagColors {
		if c == color {
			return true
		}
	}
	return false
}

// SaveTag creates or updates a tag. An empty color keeps the current color,
// or uses gray for a new tag.
func (s *Service) SaveTag(user uuid.UUID, id *uuid.UUID, name, color string) (*model.Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 40 {
		return nil, errors.New("tag name must contain 1 to 40 characters")
	}
	if color != "" && !validTagColor(color) {
		return nil, errors.New("invalid tag color")
	}
	if id == nil {
		if color == "" {
			color = "gray"
		}
		tag := model.Tag{UserID: user, Name: name, Color: color}
		err := s.DB.Create(&tag).Error
		return &tag, err
	}
	values := map[string]any{"name": name}
	if color != "" {
		values["color"] = color
	}
	r := s.DB.Model(&model.Tag{}).Where("id = ? AND user_id = ?", *id, user).Updates(values)
	if r.Error != nil {
		return nil, r.Error
	}
	if r.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	tag := model.Tag{}
	if err := s.DB.Where("id = ? AND user_id = ?", *id, user).First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

// TagsByEntity returns every tagged entity of kind with its tags, so lists can
// show tags without one request per row.
func (s *Service) TagsByEntity(user uuid.UUID, kind string) (map[uuid.UUID][]model.Tag, error) {
	if kind != "account" && kind != "activity" {
		return nil, errors.New("invalid tag target")
	}
	rows := []struct {
		model.Tag
		EntityID uuid.UUID
	}{}
	err := s.DB.Model(&model.Tag{}).Select("tags.*, a.entity_id").
		Joins("JOIN tag_assignments a ON a.tag_id = tags.id").
		Where("a.user_id = ? AND a.kind = ? AND tags.user_id = ?", user, kind, user).
		Order("tags.name").Scan(&rows).Error
	out := map[uuid.UUID][]model.Tag{}
	for _, row := range rows {
		out[row.EntityID] = append(out[row.EntityID], row.Tag)
	}
	return out, err
}
func (s *Service) DeleteTag(user, id uuid.UUID) error {
	r := s.DB.Where("id = ? AND user_id = ?", id, user).Delete(&model.Tag{})
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
func (s *Service) entityExists(db *gorm.DB, user, id uuid.UUID, kind string) error {
	var n int64
	var q *gorm.DB
	switch kind {
	case "account":
		q = db.Model(&accountmodel.Account{}).Where("id = ? AND user_id = ?", id, user)
	case "activity":
		q = db.Model(&activitymodel.Activity{}).Joins("JOIN accounts ON accounts.id = activities.account_id AND accounts.deleted_at IS NULL").Where("activities.id = ? AND accounts.user_id = ?", id, user)
	default:
		return errors.New("invalid tag target")
	}
	if err := q.Count(&n).Error; err != nil {
		return err
	}
	if n != 1 {
		return ErrNotFound
	}
	return nil
}
func (s *Service) EntityTags(user, id uuid.UUID, kind string) ([]model.Tag, error) {
	if err := s.entityExists(s.DB, user, id, kind); err != nil {
		return nil, err
	}
	rows := []model.Tag{}
	err := s.DB.Model(&model.Tag{}).Joins("JOIN tag_assignments a ON a.tag_id = tags.id").Where("a.user_id = ? AND a.entity_id = ? AND a.kind = ? AND tags.user_id = ?", user, id, kind, user).Order("tags.name").Find(&rows).Error
	return rows, err
}
func (s *Service) AssignTags(user, id uuid.UUID, kind string, ids []uuid.UUID) error {
	if len(ids) > 20 {
		return errors.New("choose at most 20 tags")
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := s.entityExists(tx, user, id, kind); err != nil {
			return err
		}
		unique := map[uuid.UUID]bool{}
		for _, tag := range ids {
			unique[tag] = true
		}
		tags := []model.Tag{}
		if len(ids) > 0 {
			if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where("id IN ? AND user_id = ?", ids, user).Find(&tags).Error; err != nil {
				return err
			}
			if len(tags) != len(unique) {
				return errors.New("tag does not belong to this user")
			}
		}
		if err := tx.Where("user_id = ? AND entity_id = ? AND kind = ?", user, id, kind).Delete(&model.TagAssignment{}).Error; err != nil {
			return err
		}
		for tag := range unique {
			if err := tx.Create(&model.TagAssignment{UserID: user, EntityID: id, Kind: kind, TagID: tag}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type ActivityFilter struct {
	AccountID                        *uuid.UUID
	TagID                            *uuid.UUID
	Query, Type, From, To, Direction string
	Page, PerPage                    int
}
type ActivityRecord struct {
	activitymodel.Activity
	AccountName string      `json:"account_name"`
	Tags        []model.Tag `gorm:"-" json:"tags"`
}
type ActivityPage struct {
	Activities []ActivityRecord `json:"activities"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PerPage    int              `json:"per_page"`
	Pages      int              `json:"pages"`
}

func (s *Service) AllActivities(user uuid.UUID, f ActivityFilter) (*ActivityPage, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Page > 100000 {
		return nil, errors.New("page is too large")
	}
	if f.PerPage < 1 {
		f.PerPage = 25
	}
	if f.PerPage > 100 {
		f.PerPage = 100
	}
	if len(f.Query) > 100 {
		return nil, errors.New("search is too long")
	}
	if f.From != "" && !validDate(f.From) || f.To != "" && !validDate(f.To) || f.From != "" && f.To != "" && f.From > f.To {
		return nil, errors.New("invalid date range")
	}
	q := s.DB.Model(&activitymodel.Activity{}).Joins("JOIN accounts ON accounts.id = activities.account_id AND accounts.deleted_at IS NULL").Where("accounts.user_id = ?", user)
	if f.AccountID != nil {
		if _, err := s.Account(user, *f.AccountID); err != nil {
			return nil, err
		}
		q = q.Where("activities.account_id = ?", *f.AccountID)
	}
	if f.TagID != nil {
		q = q.Where("EXISTS (SELECT 1 FROM tag_assignments a WHERE a.user_id = ? AND a.tag_id = ? AND ((a.kind = 'activity' AND a.entity_id = activities.id) OR (a.kind = 'account' AND a.entity_id = accounts.id)))", user, *f.TagID)
	}
	if f.Query != "" {
		value := "%" + strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(f.Query) + "%"
		q = q.Where("(activities.symbol ILIKE ? OR activities.name ILIKE ? OR activities.notes ILIKE ? OR accounts.account_name ILIKE ?)", value, value, value, value)
	}
	if f.Type != "" {
		q = q.Where("activities.type = ?", f.Type)
	}
	if f.From != "" {
		q = q.Where("activities.date >= ?", f.From)
	}
	if f.To != "" {
		q = q.Where("activities.date < CAST(? AS date) + INTERVAL '1 day'", f.To)
	}
	out := &ActivityPage{Activities: []ActivityRecord{}, Page: f.Page, PerPage: f.PerPage}
	if err := q.Count(&out.Total).Error; err != nil {
		return nil, err
	}
	out.Pages = int((out.Total + int64(f.PerPage) - 1) / int64(f.PerPage))
	if out.Pages < 1 {
		out.Pages = 1
	}
	direction := "DESC"
	if f.Direction == "asc" {
		direction = "ASC"
	}
	if err := q.Select("activities.*, accounts.account_name").Order("activities.date " + direction + ", activities.created_at " + direction + ", activities.id").Offset((f.Page - 1) * f.PerPage).Limit(f.PerPage).Find(&out.Activities).Error; err != nil {
		return nil, err
	}
	for i := range out.Activities {
		tags, err := s.EntityTags(user, out.Activities[i].ID, "activity")
		if err != nil {
			return nil, err
		}
		out.Activities[i].Tags = tags
	}
	return out, nil
}
