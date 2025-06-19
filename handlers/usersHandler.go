package handlers

import (
	"it_school/logger"
	"it_school/models"
	"it_school/repositories"
	"it_school/utils"
	"net/http"


	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserHandler struct {
	usersRepo *repositories.UsersRepository
	curatorRepo *repositories.CuratorsRepository
	roleRepo *repositories.RoleRepository
}

type CreateRequest struct {
	FullName   string    `json:"full_name"`
	Email      string    `json:"email"`
	Telephone  string    `json:"telephone"`
	Password   string    `json:"password"`
	RoleName   string 	 `json:"role_name"`
}

type CuratorResponse struct {
    ID         uuid.UUID   `json:"id"`
    FullName   string      `json:"full_name"`
    Email      string      `json:"email"`
    Telephone  string      `json:"telephone"`
    RoleID     uuid.UUID   `json:"role_id"`
    StudentIDs []uuid.UUID `json:"student_ids"`
    CourseIDs  []uuid.UUID `json:"course_ids"`
}

type UserResponse struct {
    ID                  uuid.UUID  `json:"id"`
    FullName            string     `json:"full_name"`
    Email               string     `json:"email"`
    Telephone           string     `json:"telephone"`
    RoleId              uuid.UUID  `json:"roleId"`
    RoleName            string  	`json:"role_name"`
}


func NewUserHandlers(usersRepo *repositories.UsersRepository, curatorRepo *repositories.CuratorsRepository, roleRepo *repositories.RoleRepository) *UserHandler {
	return &UserHandler{
		usersRepo: usersRepo,
		curatorRepo: curatorRepo,
		roleRepo: roleRepo,
	}
}

// FindAll godoc
// @Summary Получить список пользователей
// @Description Возвращает список всех пользователей с возможностью фильтрации по роли
// @Tags Users
// @Produce json
// @Param role query string false "Фильтр по ID роли" format(uuid)
// @Success 200 {array} models.User "Список пользователей"
// @Failure 400 {object} models.ApiError "Неверный формат UUID"
// @Failure 500 {object} models.ApiError "Ошибка сервера"
// @Router /settings/users [get]
func (h *UserHandler) FindAll(c *gin.Context) {
	logger := logger.GetLogger()

	roleParam := c.Query("role")
	var roleID *uuid.UUID

	if roleParam != "" {
		id, err := uuid.Parse(roleParam) //  правильно парсим UUID
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

// GetRole godoc
// @Summary Получить название роли по его id
// @Description Возвращает строковое представление роли по заданному UUID
// @Security ApiKeyAuth
// @Param id path string true "ID роли"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ApiError            
// @Failure 404 {object} models.ApiError             
// @Failure 500 {object} models.ApiError            
// @Router /role/{id} [get]
func (h *UserHandler) GetRole(c *gin.Context) {
    logger := logger.GetLogger()

    idParam := c.Param("id")
    if idParam == "" {
        logger.Error("Missing role id query parameter")
        c.JSON(http.StatusBadRequest, models.NewApiError("role parameter is required"))
        return
    }

    id, err := uuid.Parse(idParam)
    if err != nil {
        logger.Error("Invalid UUID format for role", zap.String("roleParam", idParam), zap.Error(err))
        c.JSON(http.StatusBadRequest, models.NewApiError("invalid role id format"))
        return
    }

    role, err := h.roleRepo.GetRoleByID(c, id)
    if err != nil {
        logger.Error("Failed to get role name from repository", zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("could not get role name"))
        return
    }

    logger.Info("Successfully retrieved role name", zap.Any("role", role))
    c.JSON(http.StatusOK, gin.H{"role": role})
}

// FindManagers godoc
// @Summary Получить список менеджеров
// @Description Возвращает список всех пользователей с ролью 'manager'
// @Tags Users
// @Produce json
// @Success 200 {array} models.User "Список менеджеров"
// @Failure 500 {object} models.ApiError "Ошибка сервера"
// @Router /settings/users/managers [get]
func (h *UserHandler) FindManagers(c *gin.Context) {
	logger := logger.GetLogger()

	manager, _ := h.roleRepo.GetRoleByName(c, "manager")
	users, err := h.usersRepo.FindAll(c.Request.Context(), &manager.Id)
	if err != nil {
		logger.Error("Failed to get managers from repository", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("could not get managers"))
		return
	}

	logger.Info("Successfully retrieved managers", zap.Int("count", len(users)))
	c.JSON(http.StatusOK, users)
}

// FindCurators godoc
// @Summary Получить список кураторов
// @Description Возвращает список всех кураторов с дополнительной информацией (студенты и курсы)
// @Tags Users
// @Produce json
// @Success 200 {array} handlers.CuratorResponse "Список кураторов с деталями"
// @Failure 500 {object} models.ApiError "Ошибка сервера"
// @Router /settings/users/curators [get]
func (h *UserHandler) FindCurators(c *gin.Context) {
	logger := logger.GetLogger()

	curatorRole, _ := h.roleRepo.GetRoleByName(c, "curator")
	users, err := h.usersRepo.FindAll(c.Request.Context(), &curatorRole.Id)
	if err != nil {
		logger.Error("Failed to get curators from repository", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("could not get curators"))
		return
	}

	var curatorsResponse []CuratorResponse
	for _, user := range users {
		curatorData, err := h.curatorRepo.GetCuratorByUserID(c.Request.Context(), user.Id)
		if err != nil {
			logger.Error("Failed to get curator data", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.NewApiError("could not get curator data"))
			return
		}

		response := CuratorResponse{
			ID:         user.Id,
			FullName:   user.Full_name,
			Email:      user.Email,
			Telephone:  user.Telephone,
			RoleID:     user.RoleID,
			StudentIDs: curatorData.StudentIDs,
			CourseIDs:  curatorData.CourseIDs,
		}
		curatorsResponse = append(curatorsResponse, response)
	}

	logger.Info("Successfully retrieved curators", zap.Int("count", len(curatorsResponse)))
	c.JSON(http.StatusOK, curatorsResponse)
}

// FindById godoc
// @Summary Получить пользователя по ID
// @Description Возвращает информацию о пользователе по его UUID
// @Tags Users
// @Produce json
// @Param userId path string true "ID пользователя" format(uuid)
// @Success 200 {object} models.User "Данные пользователя"
// @Failure 400 {object} models.ApiError "Неверный формат UUID"
// @Failure 404 {object} models.ApiError "Пользователь не найден"
// @Router /settings/users/{userId} [get]
func (h *UserHandler) FindById(c *gin.Context) {
    logger := logger.GetLogger()

    id, err := uuid.Parse(c.Param("userId"))
    if err != nil {
        logger.Error("Invalid user id", zap.String("userId", c.Param("userId")), zap.Error(err))
        c.JSON(http.StatusBadRequest, models.NewApiError("invalid user id"))
        return
    }

    user, err := h.usersRepo.FindById(c.Request.Context(), id)
    if err != nil {
        logger.Error("Failed fetching user", zap.Error(err))
        c.JSON(http.StatusNotFound, models.NewApiError("user not found"))
        return
    }

	role, err := h.roleRepo.GetRoleByID(c.Request.Context(), user.RoleID)
    if err != nil {
        logger.Error("Failed to get user role", zap.String("user_id", user.Id.String()), zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't find role"))
        return
    }

    resp := UserResponse{
        ID:        user.Id,
        FullName:  user.Full_name,
        Email:     user.Email,
        Telephone: user.Telephone,
        RoleId:    role.Id,
		RoleName:  role.Name,
    }

    c.JSON(http.StatusOK, resp)
}


// Create godoc
// @Summary Создать пользователя
// @Description Создает нового пользователя с указанной ролью. Для роли 'curator' автоматически создает связанную запись.
// @Tags Users
// @Accept json
// @Produce json
// @Param request body handlers.CreateRequest true "Данные для создания пользователя" example={"full_name": "Иванов Иван", "email": "user@example.com", "telephone": "+77071234567", "password": "securePassword123", "role_name": "curator"}
// @Success 201 {object} object{message=string} "Пользователь создан"
// @Failure 400 {object} models.ApiError "Неверные данные"
// @Failure 409 {object} models.ApiError "Email уже существует"
// @Failure 500 {object} models.ApiError "Ошибка сервера"
// @Router /settings/users [post]
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

	role, err := h.roleRepo.GetRoleByName(c, req.RoleName)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Unknown role"))
		return
	}

	newUser := models.User{
		Full_name:     req.FullName,
		Email:         req.Email,
		Telephone:     req.Telephone,
		PasswordHash:  hashedPassword,
		RoleID:        role.Id,
	}

	userID, err := h.usersRepo.Create(c, newUser)
	if err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to create user"))
		return
	}

	if role.Name == "curator" {
		// Это куратор — создаем доп. запись
		curator := models.Curator{
			UserID:     userID,
			StudentIDs: []uuid.UUID{},
			CourseIDs:  []uuid.UUID{},
		}
	
		err := h.curatorRepo.Create(c, curator)
		if err != nil {
			logger.Error("Failed to create curator data", zap.Error(err))
			c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to create curator data"))
			return
		}
	}

	logger.Info("User created successfully", zap.String("email", newUser.Email))
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

// Update godoc
// @Summary Обновить пользователя
// @Description Обновляет информацию о существующем пользователе
// @Tags Users
// @Accept json
// @Produce json
// @Param userId path string true "ID пользователя" format(uuid)
// @Param request body models.User true "Обновленные данные пользователя"
// @Success 200 "Данные обновлены"
// @Failure 400 {object} models.ApiError "Неверные данные"
// @Failure 404 {object} models.ApiError "Пользователь не найден"
// @Failure 500 {object} models.ApiError "Ошибка сервера"
// @Router /settings/users/{userId} [put]
func (h *UserHandler) Update(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("userId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error("Invalid user id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user id"))
		return
	}

	_, err = h.usersRepo.FindById(c, id)
	if err != nil {
		logger.Error("Failed to find user for update", zap.String("userID", id.String()), zap.Error(err))
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
		logger.Error("Failed to update manager", zap.String("userID", id.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to update manager"))
		return
	}

	logger.Info("User updated successfully", zap.String("userID", id.String()))
	c.Status(http.StatusOK)
}

// @Summary Обновить роль пользователя
// @Security ApiKeyAuth
// @Tags users
// @Param userId path string true "ID пользователя"
// @Param roleId query string true "ID новой роли"
// @Success 200 {object} map[string]string "role updated successfully"
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /users/{userId}/role [put]
func (h *UserHandler) UpdateUserRole(c *gin.Context) {
    logger := logger.GetLogger()

    userID, err := uuid.Parse(c.Param("userId"))
    if err != nil {
        logger.Error("Invalid user id", zap.String("userId", c.Param("userId")), zap.Error(err))
        c.JSON(http.StatusBadRequest, models.NewApiError("invalid user id"))
        return
    }

    roleIDParam := c.Query("roleId")
    if roleIDParam == "" {
        logger.Error("Missing roleId parameter")
        c.JSON(http.StatusBadRequest, models.NewApiError("roleId is required"))
        return
    }

    roleID, err := uuid.Parse(roleIDParam)
    if err != nil {
        logger.Error("Invalid role id", zap.String("roleId", roleIDParam), zap.Error(err))
        c.JSON(http.StatusBadRequest, models.NewApiError("invalid role id format"))
        return
    }

    if err := h.usersRepo.UpdateUserRole(c.Request.Context(), userID, roleID); err != nil {
        logger.Error("Failed to update user role", zap.String("userId", userID.String()), zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewApiError("could not update user role"))
        return
    }

    logger.Info("User role updated", zap.String("userId", userID.String()), zap.String("roleId", roleID.String()))
    c.JSON(http.StatusOK, gin.H{"message": "role updated successfully"})
}


// Delete godoc
// @Summary Удалить пользователя
// @Description Удаляет пользователя из системы
// @Tags Users
// @Param id path string true "ID пользователя" format(uuid)
// @Success 204 "Пользователь удален"
// @Failure 400 {object} models.ApiError "Неверный формат UUID"
// @Failure 404 {object} models.ApiError "Пользователь не найден"
// @Failure 500 {object} models.ApiError "Ошибка сервера"
// @Router /settings/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)

	if err != nil {
		logger.Error("Invalid user id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user id"))
		return
	}

	_, err = h.usersRepo.FindById(c, id)
	if err != nil {
		logger.Error("Failed to find user", zap.String("userID", id.String()), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	err = h.usersRepo.Delete(c, id)
	if err != nil {
		logger.Error("Failed to delete user", zap.String("userID", id.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	logger.Info("User deleted successfully", zap.String("userID", id.String()))
	c.Status(http.StatusOK)
}