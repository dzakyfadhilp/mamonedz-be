package handlers

import (
	"errors"
	"strconv"
	"time"

	"mamonedz/internal/models"
	"mamonedz/internal/services"
	"mamonedz/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ExpenseHandler struct {
	service  services.ExpenseService
	validate *validator.Validate
}

func NewExpenseHandler(service services.ExpenseService) *ExpenseHandler {
	v := validator.New()
	v.RegisterValidation("validcategory", func(fl validator.FieldLevel) bool {
		return models.ValidCategories[fl.Field().String()]
	})

	return &ExpenseHandler{
		service:  service,
		validate: v,
	}
}

func getUserID(c *gin.Context) uuid.UUID {
	userID, _ := c.Get("user_id")
	return *userID.(*uuid.UUID)
}

func (h *ExpenseHandler) Create(c *gin.Context) {
	var req models.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, "Validation failed: "+err.Error())
		return
	}

	userID := getUserID(c)
	expense, err := h.service.Create(userID, &req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCategory) {
			response.BadRequest(c, "Invalid category")
			return
		}
		if errors.Is(err, services.ErrInvalidDate) {
			response.BadRequest(c, "Invalid date format, use YYYY-MM-DD")
			return
		}
		response.InternalError(c, "Failed to create expense")
		return
	}

	response.Created(c, expense, "Expense created successfully")
}

func (h *ExpenseHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid expense ID")
		return
	}

	userID := getUserID(c)
	expense, err := h.service.GetByID(id, userID)
	if err != nil {
		if errors.Is(err, services.ErrExpenseNotFound) {
			response.NotFound(c, "Expense not found")
			return
		}
		response.InternalError(c, "Failed to get expense")
		return
	}

	response.Success(c, expense)
}

func (h *ExpenseHandler) GetAll(c *gin.Context) {
	userID := getUserID(c)
	filter := &models.ExpenseFilter{
		UserID: userID,
		Limit:  10,
		Offset: 0,
	}

	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			filter.StartDate = &t
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			filter.EndDate = &t
		}
	}
	if category := c.Query("category"); category != "" {
		filter.Category = &category
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	}
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	expenses, total, err := h.service.GetAll(filter)
	if err != nil {
		response.InternalError(c, "Failed to get expenses")
		return
	}

	response.SuccessWithMeta(c, expenses, &response.Meta{
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	})
}

func (h *ExpenseHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid expense ID")
		return
	}

	var req models.UpdateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, "Validation failed: "+err.Error())
		return
	}

	userID := getUserID(c)
	expense, err := h.service.Update(id, userID, &req)
	if err != nil {
		if errors.Is(err, services.ErrExpenseNotFound) {
			response.NotFound(c, "Expense not found")
			return
		}
		if errors.Is(err, services.ErrInvalidCategory) {
			response.BadRequest(c, "Invalid category")
			return
		}
		if errors.Is(err, services.ErrInvalidDate) {
			response.BadRequest(c, "Invalid date format, use YYYY-MM-DD")
			return
		}
		response.InternalError(c, "Failed to update expense")
		return
	}

	response.SuccessWithMessage(c, expense, "Expense updated successfully")
}

func (h *ExpenseHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid expense ID")
		return
	}

	userID := getUserID(c)
	if err := h.service.Delete(id, userID); err != nil {
		if errors.Is(err, services.ErrExpenseNotFound) {
			response.NotFound(c, "Expense not found")
			return
		}
		response.InternalError(c, "Failed to delete expense")
		return
	}

	response.SuccessWithMessage(c, nil, "Expense deleted successfully")
}

func (h *ExpenseHandler) GetStats(c *gin.Context) {
	period := c.DefaultQuery("period", "month")
	userID := getUserID(c)

	stats, err := h.service.GetStats(userID, period)
	if err != nil {
		response.InternalError(c, "Failed to get statistics")
		return
	}

	response.Success(c, stats)
}
