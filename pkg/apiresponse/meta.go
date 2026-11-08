package apiresponse

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Meta struct {
	RequestID  string      `json:"request_id"`
	Timestamp  time.Time   `json:"timestamp"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type requestIDKey struct{}

// RequestID returns one server-generated identity for the entire request,
// including error responses. Client-provided IDs are not trusted as log IDs.
func RequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals(requestIDKey{}).(string); ok {
		return id
	}
	id := uuid.NewString()
	c.Locals(requestIDKey{}, id)
	c.Set("X-Request-ID", id)
	return id
}

func newMeta(c *fiber.Ctx) Meta {
	return Meta{RequestID: RequestID(c), Timestamp: time.Now().UTC()}
}

// WithPagination describes an already validated, one-based result page.
// An empty collection has one empty page, matching the application's pagers.
func WithPagination(page, pageSize int, total int64) Option {
	return func(meta *Meta) {
		pages := int64(1)
		if pageSize > 0 && total > 0 {
			pages = 1 + (total-1)/int64(pageSize)
		}
		meta.Pagination = &Pagination{Page: page, PageSize: pageSize, Total: total, TotalPages: int(pages)}
	}
}
