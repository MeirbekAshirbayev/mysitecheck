package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"math-app/internal/database"
	"math-app/internal/models"

	"github.com/gin-gonic/gin"
)

// Admin API: Dashboard
func AdminDashboardHandler(c *gin.Context) {
	var lessons []models.Lesson
	database.DB.Find(&lessons)
	c.HTML(http.StatusOK, "admin_dashboard.html", gin.H{
		"Lessons": lessons,
	})
}

// Admin API: Show Add Form
func AdminAddFormHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_add.html", nil)
}

// Admin API: Process Add Form
func AdminAddHandler(c *gin.Context) {
	gradeStr := c.PostForm("grade")
	title := c.PostForm("title")
	canvaURL := c.PostForm("canva_url")
	description := c.PostForm("description")

	// Automatic Canva URL formatting
	if strings.Contains(canvaURL, "canva.com") {
		if strings.HasSuffix(canvaURL, "view") {
			canvaURL += "?embed"
		} else if strings.HasSuffix(canvaURL, "view/") {
			canvaURL = strings.TrimSuffix(canvaURL, "/") + "?embed"
		}
	}

	taskCodes := c.PostFormArray("task_codes[]")
	taskTitles := c.PostFormArray("task_titles[]")
	taskDescriptions := c.PostFormArray("task_descriptions[]")

	grade, err := strconv.Atoi(gradeStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid grade")
		return
	}

	sortOrderStr := c.PostForm("sort_order")
	sortOrder, _ := strconv.Atoi(sortOrderStr)

	lesson := models.Lesson{
		Grade:         grade,
		Title:         title,
		CanvaEmbedURL: canvaURL,
		Description:   description,
		SortOrder:     sortOrder,
	}

	for i, code := range taskCodes {
		if code != "" {
			trimmed := strings.TrimSpace(code)
			if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
				code = fmt.Sprintf(`<iframe src="%s" frameborder="0" allowfullscreen></iframe>`, trimmed)
			}

			tTitle := ""
			if i < len(taskTitles) {
				tTitle = taskTitles[i]
			}

			tDesc := ""
			if i < len(taskDescriptions) {
				tDesc = taskDescriptions[i]
			}

			lesson.Tasks = append(lesson.Tasks, models.Task{
				Title:       tTitle,
				Description: tDesc,
				Code:        code,
				Order:       i + 1,
			})
		}
	}

	if err := database.DB.Create(&lesson).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to create lesson")
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin")
}

// Admin API: Show Edit Form
func AdminEditFormHandler(c *gin.Context) {
	idParam := c.Param("id")
	var lesson models.Lesson
	if err := database.DB.Preload("Tasks").First(&lesson, idParam).Error; err != nil {
		c.String(http.StatusNotFound, "Lesson not found")
		return
	}
	c.HTML(http.StatusOK, "admin_edit.html", gin.H{
		"Lesson": lesson,
	})
}

// Admin API: Process Edit Form
func AdminEditHandler(c *gin.Context) {
	idParam := c.Param("id")
	var lesson models.Lesson
	if err := database.DB.First(&lesson, idParam).Error; err != nil {
		c.String(http.StatusNotFound, "Lesson not found")
		return
	}

	gradeStr := c.PostForm("grade")
	title := c.PostForm("title")
	canvaURL := c.PostForm("canva_url")
	description := c.PostForm("description")

	if strings.Contains(canvaURL, "canva.com") {
		if strings.HasSuffix(canvaURL, "view") {
			canvaURL += "?embed"
		} else if strings.HasSuffix(canvaURL, "view/") {
			canvaURL = strings.TrimSuffix(canvaURL, "/") + "?embed"
		}
	}

	grade, err := strconv.Atoi(gradeStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid grade")
		return
	}

	sortOrderStr := c.PostForm("sort_order")
	sortOrder, _ := strconv.Atoi(sortOrderStr)

	lesson.Grade = grade
	lesson.Title = title
	lesson.CanvaEmbedURL = canvaURL
	lesson.Description = description
	lesson.SortOrder = sortOrder

	if err := database.DB.Save(&lesson).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to update lesson")
		return
	}

	database.DB.Delete(&models.Task{}, "lesson_id = ?", lesson.ID)

	taskCodes := c.PostFormArray("task_codes[]")
	taskTitles := c.PostFormArray("task_titles[]")
	taskDescriptions := c.PostFormArray("task_descriptions[]")

	for i, code := range taskCodes {
		if code != "" {
			trimmed := strings.TrimSpace(code)
			if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
				code = fmt.Sprintf(`<iframe src="%s" frameborder="0" allowfullscreen></iframe>`, trimmed)
			}

			tTitle := ""
			if i < len(taskTitles) {
				tTitle = taskTitles[i]
			}

			tDesc := ""
			if i < len(taskDescriptions) {
				tDesc = taskDescriptions[i]
			}

			newTask := models.Task{
				LessonID:    lesson.ID,
				Title:       tTitle,
				Description: tDesc,
				Code:        code,
				Order:       i + 1,
			}
			database.DB.Create(&newTask)
		}
	}

	c.Redirect(http.StatusSeeOther, "/admin")
}

// Admin API: Delete Lesson
func AdminDeleteHandler(c *gin.Context) {
	idParam := c.Param("id")
	if err := database.DB.Delete(&models.Lesson{}, idParam).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to delete lesson")
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin")
}
