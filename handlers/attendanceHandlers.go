package handlers

import (
	"it_school/logger"
	"it_school/models"
	"it_school/repositories"
	"it_school/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

	type AttendanceHandlers struct {
		attendanceRepo *repositories.AttendanceRepository
	}

	func NewAttendanceHandlers(attendanceRepo *repositories.AttendanceRepository) *AttendanceHandlers {
		return &AttendanceHandlers{attendanceRepo: attendanceRepo}
	}

	type CreateAttendanceRequest struct {
		StudentId     uuid.UUID                   `json:"student_id" binding:"required"`
		CourseId      uuid.UUID                   `json:"course_id" binding:"required"`
		Type          string                      `json:"type" binding:"required,oneof=урок заморозка пролонгация"`
		Lesson        *AttendanceLessonInput      `json:"lesson,omitempty"`
		Freeze        *AttendanceFreezeInput      `json:"freeze,omitempty"`
		Prolongation  *AttendanceProlongationInput `json:"prolongation,omitempty"`
	}

	type AttendanceLessonInput struct {
		CuratorId     uuid.UUID  `json:"curator_id"`
		Date          string     `json:"date"`
		Format        *string    `json:"format"`
		Feedback      *string    `json:"feedback"`
		FeedbackDate  *string    `json:"feedback_date"`
		LessonStatus  string     `json:"lessons_status" binding:"required, oneof=пропущен проведен запланирован отменен"`
	}

	type AttendanceFreezeInput struct {
		StartDate string  `json:"start_date"`
		EndDate   string  `json:"end_date"`
		Comment   *string `json:"comment"`
	}

	type AttendanceProlongationInput struct {
		PaymentType string  `json:"payment_type" binding:"required,oneof=оплата предоплата доплата"`
		Date        string  `json:"date"`
		Amount      float64 `json:"amount"`
		Comment     *string `json:"comment"`
	}



// CreateAttendance godoc
// @Summary Создать запись посещаемости
// @Description Добавляет новую запись: урок, заморозку или пролонгацию
// @Description Допустимые значения:
// @Description - type: урок, заморозка, пролонгация
// @Description - lessons_status: пропущен, проведен, запланирован, отменен
// @Description - payment_type: оплата, предоплата, доплата
// @Tags Attendance
// @Accept json
// @Produce json
// @Param request body CreateAttendanceRequest true "Данные посещаемости"
// @Success 201 {object} map[string]string
// @Router /attendances [post]
func (h *AttendanceHandlers) CreateAttendance(c *gin.Context) {
	logger := logger.GetLogger()
	var req CreateAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request"))
		return
	}

	attendance := &models.Attendance{
		StudentId: req.StudentId,
		CourseId:  req.CourseId,
		Type:      req.Type,
		CreatedAt: time.Now(),
	}

	var lesson *models.AttendanceLesson
	var freeze *models.AttendanceFreeze
	var prolongation *models.AttendanceProlongation

	roleObj, _ := c.Get("userRole")
	role := roleObj.(*models.Role)

	if !utils.HasAccessToType(role, req.Type) {
		c.JSON(http.StatusForbidden, models.NewApiError("You are not allowed to create this type of attendance"))
		return
	}

	switch req.Type {
	case "урок":
		if req.Lesson == nil {
			c.JSON(http.StatusBadRequest, models.NewApiError("Lesson data is required"))
			return
		}

		parsedDate, err := utils.ParseRequiredDate(req.Lesson.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewApiError("Invalid lesson date format. Use DD.MM.YYYY"))
			return
		}

		feedbackDate, err := utils.ParseDate(req.Lesson.FeedbackDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewApiError("Invalid feedback date format. Use DD.MM.YYYY"))
			return
		}

		status := req.Lesson.LessonStatus
		if status == "" {
			if parsedDate.After(time.Now()) {
				status = "запланирован"
			} else {
				status = "проведен"
			}
		}

		lesson = &models.AttendanceLesson{
			CuratorId:    req.Lesson.CuratorId,
			Date:         parsedDate,
			Format:       req.Lesson.Format,
			Feedback:     req.Lesson.Feedback,
			LessonStatus: status,
			CreatedAt:    time.Now(),
			FeedbackDate: feedbackDate,
		}

	case "заморозка":
		if req.Freeze == nil {
			c.JSON(http.StatusBadRequest, models.NewApiError("Freeze data is required"))
			return
		}

		startDate, err := utils.ParseRequiredDate(req.Freeze.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewApiError("Invalid start date format. Use DD.MM.YYYY"))
			return
		}

		endDate, err := utils.ParseRequiredDate(req.Freeze.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewApiError("Invalid end date format. Use DD.MM.YYYY"))
			return
		}

		if startDate.After(endDate) {
			c.JSON(http.StatusBadRequest, models.NewApiError("Start date must be before end date"))
			return
		}

		freeze = &models.AttendanceFreeze{
			StartDate:    startDate,
			EndDate:      endDate,
			Comment:      req.Freeze.Comment,
		}

	case "пролонгация":
		if req.Prolongation == nil {
			c.JSON(http.StatusBadRequest, models.NewApiError("Prolongation data is required"))
			return
		}

		prolongationDate, err := utils.ParseRequiredDate(req.Prolongation.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewApiError("Invalid prolongation date format. Use DD.MM.YYYY"))
			return
		}

		prolongation = &models.AttendanceProlongation{
			PaymentType:  req.Prolongation.PaymentType,
			Date:         prolongationDate,
			Amount:       req.Prolongation.Amount,
			Comment:      req.Prolongation.Comment,
		}
	}

	id, err := h.attendanceRepo.CreateAttendance(c.Request.Context(), attendance, lesson, freeze, prolongation)
	if err != nil {
		logger.Error("Failed to create attendance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Could not create attendance"))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// GetByStudent godoc
// @Summary Получить посещаемость студента
// @Description Возвращает список всех посещений, заморозок и пролонгаций по студенту
// @Tags Attendance
// @Accept json
// @Produce json
// @Param studentId path string true "UUID студента"
// @Success 200 {array} interface{}
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /attendances/{studentId} [get]
func (h *AttendanceHandlers) GetByStudent(c *gin.Context) {
	logger := logger.GetLogger()

	studentIDStr := c.Param("student_id")
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid student UUID"))
		return
	}

	attendances, err := h.attendanceRepo.FindByStudent(c.Request.Context(), studentID)
	if err != nil {
		logger.Error("Couldn't get student's attendance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't get data"))
		return
	}

	c.JSON(http.StatusOK, attendances)
}

// UpdateAttendance godoc
// @Summary Обновить запись посещаемости
// @Description Обновляет запись посещаемости (урок, заморозка или пролонгация)
// @Description Допустимые значения:
// @Description - type: урок, заморозка, пролонгация
// @Description - lessons_status: пропущен, проведен, запланирован, отменен
// @Description - payment_type: оплата, предоплата, доплата
// @Tags Attendance
// @Accept json
// @Produce json
// @Param id path string true "ID записи посещаемости"
// @Param request body CreateAttendanceRequest true "Обновленные данные посещаемости"
// @Success 200
// @Failure 400 {object} models.ApiError
// @Failure 403 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /attendances/{id} [put]
func (h *AttendanceHandlers) UpdateAttendance(c *gin.Context) {
	logger := logger.GetLogger()

	attendanceIDParam := c.Param("id")
	attendanceID, err := uuid.Parse(attendanceIDParam)
	if err != nil {
		logger.Error("Invalid attendance ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid attendance ID"))
		return
	}

	exists, _ := h.attendanceRepo.Exists(c.Request.Context(), attendanceID)
	if !exists {
		c.JSON(http.StatusNotFound, models.NewApiError("Attendance record not found"))
		return
	}

	var req CreateAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request"))
		return
	}

	attendance := &models.Attendance{
		ID:        attendanceID,
		StudentId: req.StudentId,
		CourseId:  req.CourseId,
		Type:      req.Type,
	}

	var lesson *models.AttendanceLesson
	var freeze *models.AttendanceFreeze
	var prolongation *models.AttendanceProlongation


	roleObj, _ := c.Get("userRole")
	role := roleObj.(*models.Role)

	if !utils.HasAccessToType(role, req.Type) {
		c.JSON(http.StatusForbidden, models.NewApiError("You are not allowed to update this type of attendance"))
		return
	}

	switch req.Type {
	case "урок":
		if req.Lesson == nil {
			c.JSON(http.StatusBadRequest, models.NewApiError("Lesson data is required"))
			return
		}

		parsedDate, err := utils.ParseRequiredDate(req.Lesson.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewApiError("Invalid lesson date format. Use DD.MM.YYYY"))
			return
		}

		var feedbackDate *time.Time
		if req.Lesson.FeedbackDate != nil && *req.Lesson.FeedbackDate != "" {
			fd, err := utils.ParseDate(req.Lesson.FeedbackDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, models.NewApiError("Invalid feedback date format. Use DD.MM.YYYY"))
				return
			}
			feedbackDate = fd
		}

		status := req.Lesson.LessonStatus
		if status == "" {
			if parsedDate.After(time.Now()) {
				status = "запланирован"
			} else {
				status = "проведен"
			}
		}

		lesson = &models.AttendanceLesson{
			AttendanceID: attendanceID,
			CuratorId:    req.Lesson.CuratorId,
			Date:         parsedDate,
			Format:       req.Lesson.Format,
			Feedback:     req.Lesson.Feedback,
			LessonStatus: status,
			FeedbackDate: feedbackDate,
		}

	case "заморозка":
		if req.Freeze == nil {
			c.JSON(http.StatusBadRequest, models.NewApiError("Freeze data is required"))
			return
		}

		startDate, err := utils.ParseRequiredDate(req.Freeze.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewApiError("Invalid start date format. Use DD.MM.YYYY"))
			return
		}

		endDate, err := utils.ParseRequiredDate(req.Freeze.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewApiError("Invalid end date format. Use DD.MM.YYYY"))
			return
		}

		if startDate.After(endDate) {
			c.JSON(http.StatusBadRequest, models.NewApiError("Start date must be before end date"))
			return
		}

		freeze = &models.AttendanceFreeze{
			AttendanceID: attendanceID,
			StartDate:    startDate,
			EndDate:      endDate,
			Comment:      req.Freeze.Comment,
		}

	case "пролонгация":
		if req.Prolongation == nil {
			c.JSON(http.StatusBadRequest, models.NewApiError("Prolongation data is required"))
			return
		}

		prolongationDate, err := utils.ParseRequiredDate(req.Prolongation.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewApiError("Invalid prolongation date format. Use DD.MM.YYYY"))
			return
		}

		prolongation = &models.AttendanceProlongation{
			AttendanceID: attendanceID,
			PaymentType:  req.Prolongation.PaymentType,
			Date:         prolongationDate,
			Amount:       req.Prolongation.Amount,
			Comment:      req.Prolongation.Comment,
		}
	}

	err = h.attendanceRepo.Update(c.Request.Context(), attendance, lesson, freeze, prolongation)
	if err != nil {
		logger.Error("Failed to update attendance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Could not update attendance"))
		return
	}

	c.Status(http.StatusOK)
}

// Delete godoc
// @Summary Удалить запись посещаемости
// @Description Удаляет запись посещаемости по ID
// @Tags Attendance
// @Accept json
// @Produce json
// @Param id path string true "ID записи посещаемости"
// @Success 204
// @Failure 400 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /settings/attendance/{id} [delete]
func (h *AttendanceHandlers) Delete(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	attendanceID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid attendance UUID"))
		return
	}

	exists, _ := h.attendanceRepo.Exists(c.Request.Context(), attendanceID)
	if !exists {
		c.JSON(http.StatusNotFound, models.NewApiError("Attendance record not found"))
		return
	}

	err = h.attendanceRepo.Delete(c.Request.Context(), attendanceID)
	if err != nil {
		logger.Error("Couldn't delete attendance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't delete attendance"))
		return
	}

	c.Status(http.StatusNoContent)
}