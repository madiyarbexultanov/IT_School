package handlers

import (
	"it_school/models"
	"it_school/repositories"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CourseRequest struct {
	Title string `json:"title"`
}

type UpdateRequest struct {
	Title string `json:"title"`
}
type CourseHandlers struct {
	courseRepo *repositories.CourseRepository
}

func NewCourseHandlers(courseRepo *repositories.CourseRepository) *CourseHandlers {
	return &CourseHandlers{courseRepo: courseRepo}
}

// Create godoc
// @Summary Создать курс
// @Description Создает новый курс
// @Tags courses
// @Accept json
// @Produce json
// @Param request body CourseRequest true "Данные курса"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /courses [post]
func (h *CourseHandlers) Create(c *gin.Context) {
	var request CourseRequest
	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("couldn´t create course request"))
		return
	}

	course := models.Course{
		Title: request.Title,
	}

	id, err := h.courseRepo.Create(c, course)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

// Update godoc
// @Summary Обновить курс
// @Description Обновляет курс по ID
// @Tags courses
// @Accept json
// @Produce json
// @Param courseId path string true "ID курса"
// @Param request body UpdateRequest true "Обновлённые данные"
// @Success 200
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /courses/{courseId} [put]
func (h *CourseHandlers) Update(c *gin.Context) {
	idStr := c.Param("courseId")
	courseId, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Seasons Id"))
		return
	}

	_, err = h.courseRepo.FindById(c, courseId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var request UpdateRequest
	err = c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("couldn´t create course request"))
		return
	}

	course := models.Course{
		Id:    courseId,
		Title: request.Title,
	}

	err = h.courseRepo.Update(c, course)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

// FindById godoc
// @Summary Получить курс по ID
// @Description Возвращает курс по его идентификатору
// @Tags courses
// @Produce json
// @Param courseId path string true "ID курса"
// @Success 200 {object} models.Course
// @Failure 400 {object} models.ApiError
// @Router /courses/{courseId} [get]
func (h *CourseHandlers) FindById(c *gin.Context) {
	idStr := c.Param("courseId")
	courseId, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid coursse id"))
		return
	}

	course, err := h.courseRepo.FindById(c, courseId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, course)
}

// FindAll godoc
// @Summary Получить все курсы
// @Description Возвращает список всех курсов
// @Tags courses
// @Produce json
// @Success 200 {array} models.Course
// @Failure 400 {object} models.ApiError
// @Router /courses [get]
func (h *CourseHandlers) FindAll(c *gin.Context) {
	courses, err := h.courseRepo.FindAll(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, courses)
}

// Delete godoc
// @Summary Удалить курс
// @Description Удаляет курс по ID
// @Tags courses
// @Produce json
// @Param courseId path string true "ID курса"
// @Success 200
// @Failure 400 {object} models.ApiError
// @Router /courses/{courseId} [delete]
func (h *CourseHandlers) Delete(c *gin.Context) {
	idStr := c.Param("courseId")
	courseId, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Seasons Id"))
		return
	}

	_, err = h.courseRepo.FindById(c, courseId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	err = h.courseRepo.Delete(c, courseId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
