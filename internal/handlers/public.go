package handlers

import (
	"net/http"
	"strconv"

	"math-app/internal/database"
	"math-app/internal/models"

	"github.com/gin-gonic/gin"
)

// Home Page: Select Grade
func HomeHandler(c *gin.Context) {
	var grades []int
	// "ORDER BY grade" is better.
	database.DB.Model(&models.Lesson{}).Distinct("grade").Order("grade").Pluck("grade", &grades)

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Grades": grades,
	})
}

// Grade Page: List Topics
func GradeListHandler(c *gin.Context) {
	gradeParam := c.Param("num")
	grade, err := strconv.Atoi(gradeParam)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid grade")
		return
	}

	var lessons []models.Lesson
	database.DB.Where("grade = ?", grade).Order("sort_order ASC").Find(&lessons)

	c.HTML(http.StatusOK, "grade_list.html", gin.H{
		"Grade":   grade,
		"Lessons": lessons,
	})
}

// Lesson Page: Presentation + Tasks
func LessonHandler(c *gin.Context) {
	idParam := c.Param("id")
	var lesson models.Lesson

	// Preload Tasks
	if err := database.DB.Preload("Tasks").First(&lesson, idParam).Error; err != nil {
		c.String(http.StatusNotFound, "Lesson not found")
		return
	}

	c.HTML(http.StatusOK, "lesson.html", gin.H{
		"Lesson": lesson,
	})
}

// Task Page: Isolated View
func TaskHandler(c *gin.Context) {
	idParam := c.Param("id")
	var task models.Task
	if err := database.DB.First(&task, idParam).Error; err != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}
	c.HTML(http.StatusOK, "task.html", gin.H{
		"Task": task,
	})
}
