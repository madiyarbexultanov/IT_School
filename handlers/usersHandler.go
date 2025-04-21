package handlers

import (
	"it_school/logger"
	"it_school/models"
	"it_school/repositories"
	"it_school/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	usersRepo *repositories.UsersRepository
}

type CreateRequest struct {
	FullName 	string
	Email		string
	Telephone	string
	Password	string
	RoleID		int
}

func NewUserHandlers(usersRepo *repositories.UsersRepository) *UserHandler {
	return &UserHandler{
		usersRepo: usersRepo,
	}
}

func (h *UserHandler) FindAll(c *gin.Context) {
	logger := logger.GetLogger()

	roleParam := c.Query("role")
	var roleID *int

	if roleParam != "" {
		id, err := strconv.Atoi(roleParam)
		if err != nil {
			logger.Warn("Invalid role query param", zap.String("role", roleParam))
			c.JSON(http.StatusBadRequest, models.NewApiError("Invalid role parameter"))
			return
		}
		roleID = &id
	}

	users, err := h.usersRepo.FindAll(c.Request.Context(), roleID)
	if err != nil {
		logger.Error("Failed to get users from repository", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("could not get users"))
		return
	}

	logger.Info("Successfully retrieved users", zap.Int("count", len(users)))
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) FindManagers(c *gin.Context) {
	logger := logger.GetLogger()


	roleID := 3 
	users, err := h.usersRepo.FindAll(c.Request.Context(), &roleID)
	if err != nil {
		logger.Error("Failed to get managers from repository", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("could not get managers"))
		return
	}

	logger.Info("Successfully retrieved managers", zap.Int("count", len(users)))
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) FindCurators(c *gin.Context) {
	logger := logger.GetLogger()

	roleID := 2
	users, err := h.usersRepo.FindAll(c.Request.Context(), &roleID)
	if err != nil {
		logger.Error("Failed to get curators from repository", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("could not get curators"))
		return
	}

	logger.Info("Successfully retrieved curators", zap.Int("count", len(users)))
	c.JSON(http.StatusOK, users)
}


func (h *UserHandler) FindById(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		logger.Error("Invalid user id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user id"))
		return
	}

	user, err := h.usersRepo.FindById(c, id)
	if err != nil {
		logger.Error("Failed to find user", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("User found", zap.Int("id", id), zap.String("name", user.Full_name))
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Create(c *gin.Context) {
	logger := logger.GetLogger()

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request data", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request data"))
		return
	}

	_, err := h.usersRepo.FindByEmail(c, req.Email)
	if err == nil {
		logger.Warn("Email already exists", zap.String("email", req.Email))
		c.JSON(http.StatusConflict, models.NewApiError("Email already exists"))
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to hash password"))
		return
	}

	newUser := models.User{
		Full_name:     req.FullName,
		Email:         req.Email,
		Telephone:     req.Telephone,
		PasswordHash:  hashedPassword,
		RoleID:        req.RoleID,
	}

	_, err = h.usersRepo.Create(c, newUser)
	if err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to create user"))
		return
	}

	logger.Info("User created successfully", zap.String("email", newUser.Email))
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}


func (h *UserHandler) Update(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Invalid user id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user id"))
		return
	}

	_, err = h.usersRepo.FindById(c, id)
	if err != nil {
		logger.Error("Failed to find user for update", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request body"))
		return
	}

	err = h.usersRepo.Update(c, id, req)
	if err != nil {
		logger.Error("Failed to update manager", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to update manager"))
		return
	}

	logger.Info("Manager updated successfully", zap.Int("id", id))
	c.Status(http.StatusOK)
}


func (h *UserHandler) Delete(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		logger.Error("Invalid user id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user id"))
		return
	}

	_, err = h.usersRepo.FindById(c, id)
	if err != nil {
		logger.Error("Failed to find user", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	err = h.usersRepo.Delete(c, id)
	if err != nil {
		logger.Error("Failed to delete user", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	logger.Info("User deleted successfully", zap.Int("id", id))
	c.Status(http.StatusOK)
}