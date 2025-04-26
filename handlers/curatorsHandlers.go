package handlers

import (
	"it_school/logger"
	"it_school/models"
	"it_school/repositories"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CuratorsHandler struct {
	repo *repositories.CuratorsRepository
}

func NewCuratorsHandler(repo *repositories.CuratorsRepository) *CuratorsHandler {
	return &CuratorsHandler{repo: repo}
}

// AddStudent godoc
// @Summary Add student to curator
// @Description Assigns a student to a curator
// @Tags curators
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body handlers.CuratorsHandler.AddStudent.request true "Student assignment data"
// @Success 200 {object} object{message=string} "Student added successfully"
// @Failure 400 {object} models.ApiError "Invalid request data"
// @Failure 403 {object} models.ApiError "Forbidden"
// @Failure 404 {object} models.ApiError "Curator or student not found"
// @Failure 500 {object} models.ApiError "Internal server error"
// @Router /curators/add-student [post]
func (h *CuratorsHandler) AddStudent(c *gin.Context) {
	logger := logger.GetLogger()
	var req struct {
		CuratorID uuid.UUID `json:"curator_id"`
		StudentID uuid.UUID `json:"student_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request for adding student", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request data"))
		return
	}

	if err := h.repo.AddStudent(c, req.CuratorID, req.StudentID); err != nil {
		logger.Error("Failed to add student", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to add student"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Student added to curator"})
}

// RemoveStudent godoc
// @Summary Remove student from curator
// @Description Unassigns a student from a curator
// @Tags curators
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body handlers.CuratorsHandler.RemoveStudent.request true "Student unassignment data"
// @Success 200 {object} object{message=string} "Student removed successfully"
// @Failure 400 {object} models.ApiError "Invalid request data"
// @Failure 403 {object} models.ApiError "Forbidden"
// @Failure 404 {object} models.ApiError "Assignment not found"
// @Failure 500 {object} models.ApiError "Internal server error"
// @Router /curators/remove-student [post]
func (h *CuratorsHandler) RemoveStudent(c *gin.Context) {
	logger := logger.GetLogger()
	var req struct {
		CuratorID uuid.UUID `json:"curator_id"`
		StudentID uuid.UUID `json:"student_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request for removing student", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request data"))
		return
	}

	if err := h.repo.RemoveStudent(c, req.CuratorID, req.StudentID); err != nil {
		logger.Error("Failed to remove student", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to remove student"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Student removed from curator"})
}

// AddCourse godoc
// @Summary Add course to curator
// @Description Assigns a course to a curator
// @Tags curators
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body handlers.CuratorsHandler.AddCourse.request true "Course assignment data"
// @Success 200 {object} object{message=string} "Course added successfully"
// @Failure 400 {object} models.ApiError "Invalid request data"
// @Failure 403 {object} models.ApiError "Forbidden"
// @Failure 404 {object} models.ApiError "Curator or course not found"
// @Failure 500 {object} models.ApiError "Internal server error"
// @Router /curators/add-course [post]
func (h *CuratorsHandler) AddCourse(c *gin.Context) {
	logger := logger.GetLogger()
	var req struct {
		CuratorID uuid.UUID `json:"curator_id"`
		CourseID  uuid.UUID `json:"course_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request for adding course", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request data"))
		return
	}

	if err := h.repo.AddCourse(c, req.CuratorID, req.CourseID); err != nil {
		logger.Error("Failed to add course", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to add course"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course added to curator"})
}

// RemoveCourse godoc
// @Summary Remove course from curator
// @Description Unassigns a course from a curator
// @Tags curators
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body handlers.CuratorsHandler.RemoveCourse.request true "Course unassignment data"
// @Success 200 {object} object{message=string} "Course removed successfully"
// @Failure 400 {object} models.ApiError "Invalid request data"
// @Failure 403 {object} models.ApiError "Forbidden"
// @Failure 404 {object} models.ApiError "Assignment not found"
// @Failure 500 {object} models.ApiError "Internal server error"
// @Router /curators/remove-course [post]
func (h *CuratorsHandler) RemoveCourse(c *gin.Context) {
	logger := logger.GetLogger()
	var req struct {
		CuratorID uuid.UUID `json:"curator_id"`
		CourseID  uuid.UUID `json:"course_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request for removing course", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request data"))
		return
	}

	if err := h.repo.RemoveCourse(c, req.CuratorID, req.CourseID); err != nil {
		logger.Error("Failed to remove course", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to remove course"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course removed from curator"})
}
