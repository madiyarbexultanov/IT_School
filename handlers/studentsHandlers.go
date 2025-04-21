package handlers

import (
	"fmt"
	"net/http"
	"it_school/models"
	"it_school/repositories"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nyaruka/phonenumbers"
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

func (h *StudentsHandlers) Update(c *gin.Context) {
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
	}
	err = h.StudentsRepo.Update(c, students)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

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

func (h *StudentsHandlers) Delete(c *gin.Context) {
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

	err = h.StudentsRepo.Delete(c, studentId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
