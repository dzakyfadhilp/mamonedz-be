package models

import (
	"time"

	"github.com/google/uuid"
)

var ValidCategories = map[string]bool{
	"makanan":      true,
	"transportasi": true,
	"hiburan":      true,
	"belanja":      true,
	"kesehatan":    true,
	"pendidikan":   true,
	"lainnya":      true,
}

type Expense struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Amount    float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	Category  string    `gorm:"type:varchar(50);not null;index" json:"category"`
	Date      time.Time `gorm:"type:date;not null;index" json:"date"`
	Note      *string   `gorm:"type:text" json:"note,omitempty"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type CreateExpenseRequest struct {
	Amount   float64 `json:"amount" validate:"required,gt=0"`
	Category string  `json:"category" validate:"required,validcategory"`
	Date     string  `json:"date" validate:"required"`
	Note     *string `json:"note"`
}

type UpdateExpenseRequest struct {
	Amount   *float64 `json:"amount" validate:"omitempty,gt=0"`
	Category *string  `json:"category" validate:"omitempty,validcategory"`
	Date     *string  `json:"date"`
	Note     *string  `json:"note"`
}

type ExpenseFilter struct {
	UserID    uuid.UUID
	StartDate *time.Time
	EndDate   *time.Time
	Category  *string
	Limit     int
	Offset    int
}

type CategoryStats struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
	Count    int     `json:"count"`
}

type DailyTrend struct {
	Date  string  `json:"date"`
	Total float64 `json:"total"`
}

type ExpenseStats struct {
	Total      float64         `json:"total"`
	Count      int             `json:"count"`
	ByCategory []CategoryStats `json:"by_category"`
	DailyTrend []DailyTrend    `json:"daily_trend"`
}
