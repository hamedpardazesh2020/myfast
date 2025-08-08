package article

import "time"

// Article represents a news item stored in the database.
type Article struct {
	ID        int64
	Title     string
	URL       string
	Content   string
	ImagePath string
	VideoPath string
	Approved  bool
	Posted    bool
	CreatedAt time.Time
}
