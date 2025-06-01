package handlers

import (
	"fmt"
	"it_school/logger"
	"it_school/models"
	"it_school/repositories"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nyaruka/phonenumbers"
	"go.uber.org/zap"
)

type createStudentRequest struct {
	CourseId          uuid.UUID `json:"course_id"`
	FullName          string    `json:"full_name"`
	PhoneNumber       *string   `json:"phone_number"`
	ParentName        string    `json:"parent_name"`
	ParentPhoneNumber *string   `json:"parent_phone_number"`
	CuratorId         uuid.UUID  `json:"curator_id"`
	PlatformLink string   `json:"platform_link"`
	CrmLink      string   `json:"crm_link"`
	CreatedAt    *string  `json:"created_at"`
    IsActive     *string  `json:"is_active" enums:"активен,неактивен" example:"активен"`
}

type updateStudentRequest struct {
	CourseId          uuid.UUID `json:"course_id"`
	FullName          string    `json:"full_name"`
	PhoneNumber       *string   `json:"phone_number"`
	ParentName        string    `json:"parent_name"`
	ParentPhoneNumber *string   `json:"parent_phone_number"`
	CuratorId         uuid.UUID  `json:"curator_id"`
	PlatformLink string   `json:"platform_link"`
	CrmLink      string   `json:"crm_link"`
	CreatedAt    *string  `json:"created_at"`
	IsActive     *string  `json:"is_active" enums:"активен,неактивен" example:"активен"`
}
type StudentsHandlers struct {
	StudentsRepo *repositories.StudentsRepository
}

func NewStudentsHandlers(StudentsRepo *repositories.StudentsRepository) *StudentsHandlers {
	return &StudentsHandlers{StudentsRepo: StudentsRepo}
}

func formatPhoneNumber(input string, defaultRegion string) (string, error) {
	num, err := phonenumbers.Parse(input, defaultRegion)
	if err != nil {
		//l.Error("неверный формат номера", zap.Error(err))
		return "", fmt.Errorf("неверный формат номера: %w", err)
	}

	if !phonenumbers.IsValidNumber(num) {
		return "", fmt.Errorf("номер невалиден")
	}

	countryCode := fmt.Sprintf("+%d", num.GetCountryCode())
	nationalNumber := fmt.Sprintf("%d", num.GetNationalNumber())

	if len(nationalNumber) == 10 {
		return fmt.Sprintf("%s (%s) - %s - %s - %s",
			countryCode,
			nationalNumber[:3],
			nationalNumber[3:6],
			nationalNumber[6:8],
			nationalNumber[8:],
		), nil
	}

	return phonenumbers.Format(num, phonenumbers.INTERNATIONAL), nil
}

// Create godoc
// @Summary Создать нового студента
// @Description Создает запись о студенте. Допустимые значения:
// @Description - is_active: активен, неактивен
// @Description - created_at: дата в формате DD.MM.YYYY
// @Description - phone_number: международный формат (+7XXX...)
// @Tags Students
// @Accept json
// @Produce json
// @Param request body createStudentRequest true "Данные студента"
// @Success 201 {object} object{id=string} "ID созданного студента"
// @Failure 400 {object} models.ApiError "Неверный формат данных"
// @Failure 500 {object} models.ApiError "Ошибка сервера"
// @Router /settings/students [post]
func (h *StudentsHandlers) Create(c *gin.Context) {
    logger := logger.GetLogger()
    var request createStudentRequest
    
    if err := c.ShouldBindJSON(&request); err != nil {
        logger.Warn("Invalid student create request format", zap.Error(err))
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request data"))
        return
    }

    logger.Info("Creating student", 
        zap.String("full_name", request.FullName),
    )

    formattedPhone, err := formatPhoneNumber(*request.PhoneNumber, "KZ")
    if err != nil {
        logger.Warn("Invalid student phone format", 
            zap.String("phone", *request.PhoneNumber),
            zap.Error(err),
        )
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid student's phone number"))
        return
    }

    formattedParentsPhone, err := formatPhoneNumber(*request.ParentPhoneNumber, "KZ")
    if err != nil {
        logger.Warn("Invalid parent phone format", 
            zap.String("phone", *request.ParentPhoneNumber),
            zap.Error(err),
        )
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid parent's phone number"))
        return
    }

    CreatedAt, err := time.Parse("02.01.2006", *request.CreatedAt)
    if err != nil {
        logger.Warn("Invalid date format", 
            zap.String("date", *request.CreatedAt),
            zap.Error(err),
        )
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid created date format. Use DD.MM.YYYY"))
        return
    }

    student := models.Student{
        CourseId:          request.CourseId,
        FullName:          request.FullName,
        PhoneNumber:       &formattedPhone,
        ParentName:        request.ParentName,
        ParentPhoneNumber: &formattedParentsPhone,
        PlatformLink:      request.PlatformLink,
        CrmLink:           request.CrmLink,
        CreatedAt:         &CreatedAt,
        IsActive:          request.IsActive,
    }

    id, err := h.StudentsRepo.Create(c, student)
    if err != nil {
        logger.Error("Failed to create student", 
            zap.Error(err),
            zap.Any("student_data", student),
        )
        c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to create student"))
        return
    }

    logger.Info("Student created successfully", zap.String("student_id", id.String()))
    c.JSON(http.StatusCreated, gin.H{"id": id})
}

// FindById godoc
// @Summary Получить данные студента
// @Description Возвращает полную информацию о студенте по его ID
// @Tags Students
// @Produce json
// @Param studentId path string true "UUID студента" format(uuid)
// @Success 200 {object} models.Student "Данные студента"
// @Failure 400 {object} models.ApiError "Неверный формат UUID"
// @Failure 404 {object} models.ApiError "Студент не найден"
// @Router /settings/students/{studentId} [get]
func (h *StudentsHandlers) FindById(c *gin.Context) {
    logger := logger.GetLogger()
    idStr := c.Param("studentId")
    
    studentId, err := uuid.Parse(idStr)
    if err != nil {
        logger.Warn("Invalid student ID format", 
            zap.String("student_id", idStr),
            zap.Error(err),
        )
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid student id"))
        return
    }

    logger.Debug("Looking for student", zap.String("student_id", studentId.String()))
    
    student, err := h.StudentsRepo.FindById(c, studentId)
    if err != nil {
        logger.Error("Student not found", 
            zap.String("student_id", studentId.String()),
            zap.Error(err),
        )
        c.JSON(http.StatusNotFound, models.NewApiError("Student not found"))
        return
    }

    logger.Debug("Student found", zap.String("student_id", studentId.String()))
    c.JSON(http.StatusOK, student)
}

// Update godoc
// @Summary Обновить данные студента
// @Description Обновляет информацию о существующем студенте. Допустимые значения:
// @Description - is_active: активен, неактивен
// @Description - created_at: дата в формате DD.MM.YYYY
// @Description - phone_number: международный формат (+7XXX...)
// @Tags Students
// @Accept json
// @Produce json
// @Param studentId path string true "UUID студента" format(uuid)
// @Param is_active body string true "Статус студента" Enums(активен, неактивен)
// @Param request body updateStudentRequest true "Обновленные данные"
// @Success 200 "Данные успешно обновлены"
// @Failure 400 {object} models.ApiError "Неверный формат данных"
// @Failure 404 {object} models.ApiError "Студент не найден"
// @Failure 500 {object} models.ApiError "Ошибка сервера"
// @Router /settings/students/{studentId} [put]
func (h *StudentsHandlers) Update(c *gin.Context) {
    logger := logger.GetLogger()
    idStr := c.Param("studentId")
    
    studentId, err := uuid.Parse(idStr)
    if err != nil {
        logger.Warn("Invalid student ID format", 
            zap.String("input_id", idStr),
            zap.Error(err),
        )
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid student Id"))
        return
    }

    logger.Info("Updating student", zap.String("student_id", studentId.String()))

    var request updateStudentRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        logger.Warn("Invalid update request format", 
            zap.Error(err),
            zap.String("student_id", studentId.String()),
        )
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request payload"))
        return
    }

    formattedPhone, err := formatPhoneNumber(*request.PhoneNumber, "KZ")
    if err != nil {
        logger.Warn("Invalid student phone format in update", 
            zap.String("phone", *request.PhoneNumber),
            zap.Error(err),
        )
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid student's phone number"))
        return
    }

    formattedParentsPhone, err := formatPhoneNumber(*request.ParentPhoneNumber, "KZ")
    if err != nil {
        logger.Warn("Invalid parent phone format in update", 
            zap.String("phone", *request.ParentPhoneNumber),
            zap.Error(err),
        )
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid parent's phone number"))
        return
    }

    CreatedAt, err := time.Parse("02.01.2006", *request.CreatedAt)
    if err != nil {
        logger.Warn("Invalid date format in update", 
            zap.String("date", *request.CreatedAt),
            zap.Error(err),
        )
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid created date format. Use DD.MM.YYYY"))
        return
    }

    student := models.Student{
        Id:                studentId,
        CourseId:          request.CourseId,
        FullName:          request.FullName,
        PhoneNumber:       &formattedPhone,
        ParentName:        request.ParentName,
        ParentPhoneNumber: &formattedParentsPhone,
        PlatformLink:      request.PlatformLink,
        CrmLink:           request.CrmLink,
        CreatedAt:         &CreatedAt,
        IsActive:          request.IsActive,
    }

    if err := h.StudentsRepo.Update(c, student); err != nil {
        logger.Error("Failed to update student", 
            zap.String("student_id", studentId.String()),
            zap.Error(err),
            zap.Any("update_data", student),
        )
        c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to update student"))
        return
    }

    logger.Info("Student updated successfully", zap.String("student_id", studentId.String()))
    c.Status(http.StatusOK)
}

// FindAll godoc
// @Summary Получить список студентов
// @Description Возвращает список студентов с возможностью фильтрации
// @Tags Students
// @Produce json
// @Param search query string false "Поиск по ФИО"
// @Param course query string false "Фильтр по ID курса" format(uuid)
// @Param is_active body string false "Фильтр по активности" Enums(активен, неактивен)
// @Param curator_id query string false "Фильтр по ID куратора" format(uuid)
// @Success 200 {array} models.Student "Список студентов"
// @Failure 500 {object} models.ApiError "Ошибка сервера"
// @Router /settings/students [get]
func (h *StudentsHandlers) FindAll(c *gin.Context) {
    logger := logger.GetLogger()

    filters := models.StudentFilters{
        Search:    c.Query("search"),
        Course:    c.Query("course"),
        IsActive:  c.Query("is_active"),
        CuratorId: c.Query("curator_id"),
    }

    logger.Debug("Fetching students with filters", 
        zap.Any("filters", filters),
    )

    students, err := h.StudentsRepo.FindAll(c, filters)
    if err != nil {
        logger.Error("Failed to fetch students", 
            zap.Error(err),
            zap.Any("filters", filters),
        )
        c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to fetch students"))
        return
    }

    logger.Debug("Students fetched successfully", 
        zap.Int("count", len(students)),
    )
    c.JSON(http.StatusOK, students)
}


// Delete godoc
// @Summary Удалить студента
// @Description Удаляет запись о студенте из системы
// @Tags Students
// @Param studentId path string true "UUID студента" format(uuid)
// @Success 204 "Студент успешно удален"
// @Failure 400 {object} models.ApiError "Неверный формат UUID"
// @Failure 404 {object} models.ApiError "Студент не найден"
// @Failure 500 {object} models.ApiError "Ошибка сервера"
// @Router /settings/students/{studentId} [delete]
func (h *StudentsHandlers) Delete(c *gin.Context) {
    logger := logger.GetLogger()
    idStr := c.Param("studentId")
    
    studentId, err := uuid.Parse(idStr)
    if err != nil {
        logger.Warn("Invalid student ID format", 
            zap.String("input_id", idStr),
            zap.Error(err),
        )
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid student id"))
        return
    }

    logger.Info("Deleting student", zap.String("student_id", studentId.String()))

    if _, err := h.StudentsRepo.FindById(c, studentId); err != nil {
        logger.Warn("Student not found for deletion", 
            zap.String("student_id", studentId.String()),
            zap.Error(err),
        )
        c.JSON(http.StatusNotFound, models.NewApiError("Student not found"))
        return
    }

    if err := h.StudentsRepo.Delete(c, studentId); err != nil {
        logger.Error("Failed to delete student", 
            zap.String("student_id", studentId.String()),
            zap.Error(err),
        )
        c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to delete student"))
        return
    }

    logger.Info("Student deleted successfully", zap.String("student_id", studentId.String()))
    c.Status(http.StatusOK)
}