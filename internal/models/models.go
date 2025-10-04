package models

type URL struct {
	ID          string `db:"id"`
	UserID      string `db:"user_id"`
	ShortURL    string `db:"short_url"`
	OriginalURL string `db:"original_url"`
	IsDeleted   bool   `json:"is_deleted" db:"is_deleted"`
}

type BatchRequest []BatchRequestItem

type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse []BatchResponseItem

type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
