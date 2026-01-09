package handlers

import (
	"errors"

	"mamonedz/internal/models"
	"mamonedz/internal/services"
	"mamonedz/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	service  services.AuthService
	validate *validator.Validate
}

func NewAuthHandler(service services.AuthService) *AuthHandler {
	return &AuthHandler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, "Validation failed: "+err.Error())
		return
	}

	result, err := h.service.Register(&req)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyExists) {
			response.BadRequest(c, "Email already exists")
			return
		}
		response.InternalError(c, "Failed to register user")
		return
	}

	response.Created(c, result, "User registered successfully")
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, "Validation failed: "+err.Error())
		return
	}

	result, err := h.service.Login(&req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			response.Error(c, 401, "Invalid email or password")
			return
		}
		response.InternalError(c, "Failed to login")
		return
	}

	response.SuccessWithMessage(c, result, "Login successful")
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		response.Error(c, 401, "Unauthorized")
		return
	}

	response.Success(c, user.(*models.User).ToResponse())
}
