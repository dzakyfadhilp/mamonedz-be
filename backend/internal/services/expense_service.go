package services

import (
	"errors"
	"time"

	"mamonedz/internal/models"
	"mamonedz/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrExpenseNotFound = errors.New("expense not found")
	ErrInvalidCategory = errors.New("invalid category")
	ErrInvalidDate     = errors.New("invalid date format, use YYYY-MM-DD")
)

type ExpenseService interface {
	Create(userID uuid.UUID, req *models.CreateExpenseRequest) (*models.Expense, error)
	GetByID(id, userID uuid.UUID) (*models.Expense, error)
	GetAll(filter *models.ExpenseFilter) ([]models.Expense, int64, error)
	Update(id, userID uuid.UUID, req *models.UpdateExpenseRequest) (*models.Expense, error)
	Delete(id, userID uuid.UUID) error
	GetStats(userID uuid.UUID, period string) (*models.ExpenseStats, error)
}

type expenseService struct {
	repo repository.ExpenseRepository
}

func NewExpenseService(repo repository.ExpenseRepository) ExpenseService {
	return &expenseService{repo: repo}
}

func (s *expenseService) Create(userID uuid.UUID, req *models.CreateExpenseRequest) (*models.Expense, error) {
	if !models.ValidCategories[req.Category] {
		return nil, ErrInvalidCategory
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, ErrInvalidDate
	}

	expense := &models.Expense{
		UserID:   userID,
		Amount:   req.Amount,
		Category: req.Category,
		Date:     date,
		Note:     req.Note,
	}

	if err := s.repo.Create(expense); err != nil {
		return nil, err
	}

	return expense, nil
}

func (s *expenseService) GetByID(id, userID uuid.UUID) (*models.Expense, error) {
	expense, err := s.repo.GetByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExpenseNotFound
		}
		return nil, err
	}
	return expense, nil
}

func (s *expenseService) GetAll(filter *models.ExpenseFilter) ([]models.Expense, int64, error) {
	return s.repo.GetAll(filter)
}

func (s *expenseService) Update(id, userID uuid.UUID, req *models.UpdateExpenseRequest) (*models.Expense, error) {
	expense, err := s.repo.GetByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExpenseNotFound
		}
		return nil, err
	}

	if req.Amount != nil {
		expense.Amount = *req.Amount
	}
	if req.Category != nil {
		if !models.ValidCategories[*req.Category] {
			return nil, ErrInvalidCategory
		}
		expense.Category = *req.Category
	}
	if req.Date != nil {
		date, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			return nil, ErrInvalidDate
		}
		expense.Date = date
	}
	if req.Note != nil {
		expense.Note = req.Note
	}

	expense.UpdatedAt = time.Now()

	if err := s.repo.Update(expense); err != nil {
		return nil, err
	}

	return expense, nil
}

func (s *expenseService) Delete(id, userID uuid.UUID) error {
	_, err := s.repo.GetByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrExpenseNotFound
		}
		return err
	}
	return s.repo.Delete(id, userID)
}

func (s *expenseService) GetStats(userID uuid.UUID, period string) (*models.ExpenseStats, error) {
	now := time.Now()
	var startDate, endDate time.Time

	switch period {
	case "day":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 0, 1).Add(-time.Second)
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		startDate = time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 0, 7).Add(-time.Second)
	case "month":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, 0).Add(-time.Second)
	default:
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, 0).Add(-time.Second)
	}

	return s.repo.GetStats(userID, &startDate, &endDate)
}
