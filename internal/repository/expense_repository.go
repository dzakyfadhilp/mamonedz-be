package repository

import (
	"time"

	"mamonedz/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExpenseRepository interface {
	Create(expense *models.Expense) error
	GetByID(id, userID uuid.UUID) (*models.Expense, error)
	GetAll(filter *models.ExpenseFilter) ([]models.Expense, int64, error)
	Update(expense *models.Expense) error
	Delete(id, userID uuid.UUID) error
	GetStats(userID uuid.UUID, startDate, endDate *time.Time) (*models.ExpenseStats, error)
}

type expenseRepository struct {
	db *gorm.DB
}

func NewExpenseRepository(db *gorm.DB) ExpenseRepository {
	return &expenseRepository{db: db}
}

func (r *expenseRepository) Create(expense *models.Expense) error {
	return r.db.Create(expense).Error
}

func (r *expenseRepository) GetByID(id, userID uuid.UUID) (*models.Expense, error) {
	var expense models.Expense
	err := r.db.First(&expense, "id = ? AND user_id = ?", id, userID).Error
	if err != nil {
		return nil, err
	}
	return &expense, nil
}

func (r *expenseRepository) GetAll(filter *models.ExpenseFilter) ([]models.Expense, int64, error) {
	var expenses []models.Expense
	var total int64

	query := r.db.Model(&models.Expense{}).Where("user_id = ?", filter.UserID)

	if filter.StartDate != nil {
		query = query.Where("date >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("date <= ?", filter.EndDate)
	}
	if filter.Category != nil && *filter.Category != "" {
		query = query.Where("category = ?", *filter.Category)
	}

	query.Count(&total)

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err := query.Order("date DESC, created_at DESC").Find(&expenses).Error
	return expenses, total, err
}

func (r *expenseRepository) Update(expense *models.Expense) error {
	return r.db.Save(expense).Error
}

func (r *expenseRepository) Delete(id, userID uuid.UUID) error {
	return r.db.Delete(&models.Expense{}, "id = ? AND user_id = ?", id, userID).Error
}

func (r *expenseRepository) GetStats(userID uuid.UUID, startDate, endDate *time.Time) (*models.ExpenseStats, error) {
	stats := &models.ExpenseStats{}

	query := r.db.Model(&models.Expense{}).Where("user_id = ?", userID)
	if startDate != nil {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("date <= ?", endDate)
	}

	var result struct {
		Total float64
		Count int
	}
	query.Select("COALESCE(SUM(amount), 0) as total, COUNT(*) as count").Scan(&result)
	stats.Total = result.Total
	stats.Count = result.Count

	var categoryStats []models.CategoryStats
	catQuery := r.db.Model(&models.Expense{}).Where("user_id = ?", userID)
	if startDate != nil {
		catQuery = catQuery.Where("date >= ?", startDate)
	}
	if endDate != nil {
		catQuery = catQuery.Where("date <= ?", endDate)
	}
	catQuery.Select("category, COALESCE(SUM(amount), 0) as total, COUNT(*) as count").
		Group("category").
		Order("total DESC").
		Scan(&categoryStats)
	stats.ByCategory = categoryStats

	var dailyTrend []models.DailyTrend
	trendQuery := r.db.Model(&models.Expense{}).Where("user_id = ?", userID)
	if startDate != nil {
		trendQuery = trendQuery.Where("date >= ?", startDate)
	}
	if endDate != nil {
		trendQuery = trendQuery.Where("date <= ?", endDate)
	}
	trendQuery.Select("TO_CHAR(date, 'YYYY-MM-DD') as date, COALESCE(SUM(amount), 0) as total").
		Group("date").
		Order("date ASC").
		Scan(&dailyTrend)
	stats.DailyTrend = dailyTrend

	return stats, nil
}
