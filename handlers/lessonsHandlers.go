package handlers

import (
	"it_school/models"
	"it_school/repositories"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createLessonsrequest struct {
	StudentId uuid.UUID `json:"student_id"`
	CourseId  uuid.UUID `json:"course_id"`
	Date      *string   `json:"date"`
	Feedback  string    `json:"feedback"`
	// PaymentStatus string    `json:"payment_status"`
	// LessonsStatus string    `json:"lessons_status"`
	FeedbackDate *string `json:"feedback_date"`
	CreatedAt    *string `json:"created_at"`
}

type updateLessonsrequest struct {
	StudentId     uuid.UUID `json:"student_id"`
	CourseId      uuid.UUID `json:"course_id"`
	Date          *string   `json:"date"`
	Feedback      string    `json:"feedback"`
	PaymentStatus string    `json:"payment_status"`
	LessonsStatus string    `json:"lessons_status"`
	FeedbackDate  *string   `json:"feedback_date"`
	CreatedAt     *string   `json:"created_at"`
}

type LessonsHandlers struct {
	LessonsRepo *repositories.LessonsRepository
}

func NewLessonsHandlers(LessonsRepo *repositories.LessonsRepository) *LessonsHandlers {
	return &LessonsHandlers{
		LessonsRepo: LessonsRepo,
	}
}

func (h *LessonsHandlers) Create(c *gin.Context) {
	var request createLessonsrequest
	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("couldn´t create lessons request"))
		return
	}

	date, err := time.Parse("02.01.2006", *request.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid created date format. Use DD.MM.YYYY"))
		return
	}

	feedbackdate, err := time.Parse("02.01.2006", *request.FeedbackDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid created feedbackdate format. Use DD.MM.YYYY"))
		return
	}

	createdAt, err := time.Parse("02.01.2006", *request.CreatedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid created createdAt format. Use DD.MM.YYYY"))
		return
	}

	lessons := models.Lessons{
		StudentId: request.StudentId,
		CourseId:  request.CourseId,
		Date:      &date,
		Feedback:  request.Feedback,
		// PaymentStatus: request.PaymentStatus,
		// LessonsStatus: request.LessonsStatus,
		FeedbackDate: &feedbackdate,
		CreatedAt:    &createdAt,
	}

	id, err := h.LessonsRepo.Create(c, lessons)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

func (h *LessonsHandlers) FindById(c *gin.Context) {
	idStr := c.Param("lessonsId")
	lessonsId, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid students id"))
		return
	}

	lessons, err := h.LessonsRepo.FindById(c, lessonsId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, lessons)
}

func (h *LessonsHandlers) FindAll(c *gin.Context) {
	filters := models.LessonsFilters{
		PaymentStatus: c.Query("payment_status"),
		LessonsStatus: c.Query("lessons_status"),
	}
	lessons, err := h.LessonsRepo.FindAll(c, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, lessons)
}

func (h *LessonsHandlers) Update(c *gin.Context) {
	idStr := c.Param("lessonsId")
	lessonsId, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid students id"))
		return
	}

	lessons, err := h.LessonsRepo.FindById(c, lessonsId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var request updateLessonsrequest
	err = c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("couldn´t create lessons request"))
		return
	}

	date, err := time.Parse("02.01.2006", *request.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid created date format. Use DD.MM.YYYY"))
		return
	}

	feedbackdate, err := time.Parse("02.01.2006", *request.FeedbackDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid created date format. Use DD.MM.YYYY"))
		return
	}

	createdAt, err := time.Parse("02.01.2006", *request.CreatedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid created date format. Use DD.MM.YYYY"))
		return
	}

	lessons = models.Lessons{
		Id:            lessonsId,
		StudentId:     request.StudentId,
		CourseId:      request.CourseId,
		Date:          &date,
		Feedback:      request.Feedback,
		PaymentStatus: &request.PaymentStatus,
		LessonsStatus: &request.LessonsStatus,
		FeedbackDate:  &feedbackdate,
		CreatedAt:     &createdAt,
	}

	err = h.LessonsRepo.Update(c, lessons)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

func (h *LessonsHandlers) Delete(c *gin.Context) {
	idStr := c.Param("lessonsId")
	lessonsId, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid students id"))
		return
	}

	_, err = h.LessonsRepo.FindById(c, lessonsId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}
	err = h.LessonsRepo.Delete(c, lessonsId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
