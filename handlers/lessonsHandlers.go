package handlers

import (
	"it_school/models"
	"it_school/repositories"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Упрощенная структура запроса - оставляем только то, что действительно нужно передавать
type createLessonsrequest struct {
	StudentId uuid.UUID `json:"student_id"`
	CourseId  uuid.UUID `json:"course_id"`
	Date      *string   `json:"date"` // Опционально - если не указана, берем текущую дату
	Feedback  string    `json:"feedback"`
}

type updateLessonsrequest struct {
	StudentId     uuid.UUID `json:"student_id"`
	CourseId      uuid.UUID `json:"course_id"`
	Date          *string   `json:"date"`
	Feedback      string    `json:"feedback"`
	PaymentStatus string    `json:"payment_status"`
	LessonsStatus string    `json:"lessons_status"`
}

type LessonsHandlers struct {
	LessonsRepo *repositories.LessonsRepository
}

func NewLessonsHandlers(LessonsRepo *repositories.LessonsRepository) *LessonsHandlers {
	return &LessonsHandlers{
		LessonsRepo: LessonsRepo,
	}
}

// Create godoc
// @Summary Создать урок
// @Description Создает новый урок для студента по курсу
// @Tags lessons
// @Accept json
// @Produce json
// @Param request body createLessonsrequest true "Данные урока"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /lessons [post]
func (h *LessonsHandlers) Create(c *gin.Context) {
	var request createLessonsrequest
	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("couldn't create lessons request"))
		return
	}

	now := time.Now()
	var lessonDate time.Time
	var lessonStatus string

	// Обработка даты урока
	if request.Date != nil {
		// Если дата передана - парсим ее
		parsedDate, err := time.Parse("02.01.2006", *request.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewApiError("Invalid date format. Use DD.MM.YYYY"))
			return
		}
		lessonDate = parsedDate
		
		// Определяем статус на основе даты
		if lessonDate.After(now) {
			lessonStatus = "запланирован"
		} else {
			lessonStatus = "проведен"
		}
	} else {
		// Если дата не передана - используем текущую дату и статус "проведен"
		lessonDate = now
		lessonStatus = "проведен"
	}

	// Создаем указатели для строковых значений
	defaultPaymentStatus := "не оплачено"
	paymentStatusPtr := &defaultPaymentStatus
	lessonStatusPtr := &lessonStatus

	// Создаем объект урока с автоматически заполненными полями
	lessons := models.Lessons{
		StudentId:     request.StudentId,
		CourseId:      request.CourseId,
		Date:          &lessonDate,
		Feedback:      request.Feedback,
		PaymentStatus: paymentStatusPtr, // Теперь это *string
		LessonsStatus: lessonStatusPtr,  // Указатель на статус
		FeedbackDate:  nil,              // Дата отзыва пока не установлена
		CreatedAt:     &now,
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

// Create godoc
// @Summary Создать урок
// @Description Создает новый урок для студента по курсу
// @Tags lessons
// @Accept json
// @Produce json
// @Param request body createLessonsrequest true "Данные урока"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /lessons [post]
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

// FindAll godoc
// @Summary Получить все уроки
// @Description Возвращает список всех уроков с фильтрацией
// @Tags lessons
// @Produce json
// @Param payment_status query string false "Статус оплаты"
// @Param lessons_status query string false "Статус урока"
// @Success 200 {array} models.Lessons
// @Failure 500 {object} models.ApiError
// @Router /lessons [get]
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

// Update godoc
// @Summary Обновить урок
// @Description Обновляет данные урока по ID
// @Tags lessons
// @Accept json
// @Produce json
// @Param lessonsId path string true "ID урока"
// @Param request body updateLessonsrequest true "Обновлённые данные урока"
// @Success 200
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /lessons/{lessonsId} [put]
func (h *LessonsHandlers) Update(c *gin.Context) {
	idStr := c.Param("lessonsId")
	lessonsId, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid lesson id"))
		return
	}

	// Получаем текущий урок
	existingLesson, err := h.LessonsRepo.FindById(c, lessonsId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var request updateLessonsrequest
	if err := c.Bind(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("couldn't parse lesson request"))
		return
	}

	// Обработка даты урока
	var lessonDate time.Time
	if request.Date != nil {
		parsedDate, err := time.Parse("02.01.2006", *request.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewApiError("Invalid date format. Use DD.MM.YYYY"))
			return
		}
		lessonDate = parsedDate
	} else if existingLesson.Date != nil {
		lessonDate = *existingLesson.Date
	} else {
		lessonDate = time.Now()
	}

	// Обработка статусов
	paymentStatus := request.PaymentStatus
	if paymentStatus == "" && existingLesson.PaymentStatus != nil {
		paymentStatus = *existingLesson.PaymentStatus
	} else if paymentStatus == "" {
		paymentStatus = "не оплачено"
	}

	lessonStatus := request.LessonsStatus
	if lessonStatus == "" && existingLesson.LessonsStatus != nil {
		lessonStatus = *existingLesson.LessonsStatus
	} else if lessonStatus == "" {
		// Автоматически определяем статус на основе даты
		if lessonDate.After(time.Now()) {
			lessonStatus = "запланирован"
		} else {
			lessonStatus = "проведен"
		}
	}

	// Обработка даты отзыва
	var feedbackDate *time.Time
	if request.Feedback != "" && existingLesson.FeedbackDate == nil {
		now := time.Now()
		feedbackDate = &now
	} else {
		feedbackDate = existingLesson.FeedbackDate
	}

	// Создаем указатели для строковых значений
	paymentStatusPtr := &paymentStatus
	lessonStatusPtr := &lessonStatus

	updatedLesson := models.Lessons{
		Id:            lessonsId,
		StudentId:     request.StudentId,
		CourseId:      request.CourseId,
		Date:          &lessonDate,
		Feedback:      request.Feedback,
		PaymentStatus: paymentStatusPtr,
		LessonsStatus: lessonStatusPtr,
		FeedbackDate:  feedbackDate,
		CreatedAt:     existingLesson.CreatedAt, // Дата создания не меняется
	}

	if err := h.LessonsRepo.Update(c, updatedLesson); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

// Delete godoc
// @Summary Удалить урок
// @Description Удаляет урок по ID
// @Tags lessons
// @Produce json
// @Param lessonsId path string true "ID урока"
// @Success 200
// @Failure 400 {object} models.ApiError
// @Router /lessons/{lessonsId} [delete]
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
