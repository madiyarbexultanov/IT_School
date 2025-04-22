package handlers

import (
	"fmt"
	"it_school/logger"
	"it_school/models"
	"it_school/repositories"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nyaruka/phonenumbers"
	"go.uber.org/zap"
)

type createStudentRequest struct {
	FullName          string  `json:"full_name"`
	PhoneNumber       *string `json:"phone_number"`
	ParentName        string  `json:"parent_name"`
	ParentPhoneNumber *string `json:"parent_phone_number"`
	//CuratorId         uuid.UUID  `json:"curator_id"`
	Courses      []string `json:"courses"`
	PlatformLink string   `json:"platform_link"`
	CrmLink      string   `json:"crm_link"`
	CreatedAt    *string  `json:"created_at"`
}

type updateStudentRequest struct {
	FullName          string  `json:"full_name"`
	PhoneNumber       *string `json:"phone_number"`
	ParentName        string  `json:"parent_name"`
	ParentPhoneNumber *string `json:"parent_phone_number"`
	//CuratorId         uuid.UUID  `json:"curator_id"`
	Courses      []string `json:"courses"`
	PlatformLink string   `json:"platform_link"`
	CrmLink      string   `json:"crm_link"`
	CreatedAt    *string  `json:"created_at"`
	IsActive     *string  `json:"is_active"`
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

// @Summary Create new students
// @Description Создание студента с курсами, номерами телефонов и ссылками
// @Tags Students
// @Accept json
// @Produce json
// @Param student body createStudentRequest true "Информация о студенте"
// @Success 200 {object} map[string]string "id студента"
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /students [post]
func (h *StudentsHandlers) Create(c *gin.Context) {
	var request createStudentRequest
	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("couldn´t create student request"))
		return
	}

	formattedPhone, err := formatPhoneNumber(*request.PhoneNumber, "KZ")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid student's phone number"))
		return
	}

	formattedParentsPhone, err := formatPhoneNumber(*request.ParentPhoneNumber, "KZ")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid parent's phone number"))
		return
	}

	CreatedAt, err := time.Parse("02.01.2006", *request.CreatedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid created date format. Use DD.MM.YYYY"))
		return
	}

	students := models.Student{
		FullName:          request.FullName,
		PhoneNumber:       &formattedPhone,
		ParentName:        request.ParentName,
		ParentPhoneNumber: &formattedParentsPhone,
		Courses:           request.Courses,
		PlatformLink:      request.PlatformLink,
		CrmLink:           request.CrmLink,
		CreatedAt:         &CreatedAt,
	}

	id, err := h.StudentsRepo.Create(c, students)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

// @Summary Найти студента по ID
// @Description Получение информации о студенте по UUID
// @Tags Students
// @Produce json
// @Param studentId path string true "ID студента"
// @Success 200 {object} models.Student
// @Failure 400 {object} models.ApiError
// @Router /students/{studentId} [get]
func (h *StudentsHandlers) FindById(c *gin.Context) {
	idStr := c.Param("studentId")
	studentId, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid students id"))
		return
	}

	Students, err := h.StudentsRepo.FindById(c, studentId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, Students)
}

// @Summary Обновить данные студента
// @Description Обновление информации о студенте по UUID
// @Tags Students
// @Accept json
// @Param studentId path string true "ID студента"
// @Param student body updateStudentRequest true "Обновлённые данные студента"
// @Success 200
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /students/{studentId} [put]
func (h *StudentsHandlers) Update(c *gin.Context) {
	l := logger.GetLogger()
	idStr := c.Param("studentId")
	studentId, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Seasons Id"))
		return
	}

	_, err = h.StudentsRepo.FindById(c, studentId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var request updateStudentRequest
	err = c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request payload")
		return
	}

	formattedPhone, err := formatPhoneNumber(*request.PhoneNumber, "KZ")
	if err != nil {
		log.Println("Error formatting phone number:", err)
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid student's phone number"))
		return
	}

	formattedParentsPhone, err := formatPhoneNumber(*request.ParentPhoneNumber, "KZ")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid parent's phone number"))
		return
	}

	CreatedAt, err := time.Parse("02.01.2006", *request.CreatedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid created date format. Use DD.MM.YYYY"))
		return
	}
	students := models.Student{
		Id:                studentId,
		FullName:          request.FullName,
		PhoneNumber:       &formattedPhone,
		ParentName:        request.ParentName,
		ParentPhoneNumber: &formattedParentsPhone,
		Courses:           request.Courses,
		PlatformLink:      request.PlatformLink,
		CrmLink:           request.CrmLink,
		CreatedAt:         &CreatedAt,
		IsActive:          request.IsActive,
	}
	err = h.StudentsRepo.Update(c, students)
	if err != nil {
		l.Error("Ошибка при обновлении студента", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to update student"))
		return
	}
	c.Status(http.StatusOK)
}

// @Summary Получить список студентов
// @Description Список студентов с фильтрами по имени, курсу, активности и куратору
// @Tags Students
// @Produce json
// @Param search query string false "Поиск по имени"
// @Param course query string false "Фильтр по курсу"
// @Param is_active query string false "Фильтр по активности"
// @Param curator_id query string false "ID куратора"
// @Success 200 {array} models.Student
// @Failure 500
// @Router /students [get]
func (h *StudentsHandlers) FindAll(c *gin.Context) {

	filters := models.StudentFilters{
		Search:    c.Query("search"),
		Course:    c.Query("course"),
		IsActive:  c.Query("is_active"),
		CuratorId: c.Query("curator_id"),
	}

	Seasons, err := h.StudentsRepo.FindAll(c, filters)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, Seasons)
}

// @Summary Удалить студента
// @Description Удаление студента по UUID
// @Tags Students
// @Param studentId path string true "ID студента"
// @Success 200
// @Failure 400 {object} models.ApiError
// @Router /students/{studentId} [delete]
func (h *StudentsHandlers) Delete(c *gin.Context) {
	idStr := c.Param("studentId")
	studentId, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid students Id"))
		return
	}

	_, err = h.StudentsRepo.FindById(c, studentId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	err = h.StudentsRepo.Delete(c, studentId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
