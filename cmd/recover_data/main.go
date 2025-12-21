package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Lesson struct {
	ID            uint   `gorm:"primaryKey"`
	Grade         int    `gorm:"not null"`
	Title         string `gorm:"not null"`
	CanvaEmbedURL string `gorm:"not null"`
	Description   string
	SortOrder     int `gorm:"default:0"`
}

type Task struct {
	ID          uint `gorm:"primaryKey"`
	LessonID    uint `gorm:"not null"`
	Title       string
	Description string
	Code        string `gorm:"type:text;not null"`
	Order       int
}

type LessonInfo struct {
	Grade       int
	Title       string
	CanvaURL    string
	Description string
}

type TaskInfo struct {
	Title       string
	Description string
	Code        string
}

func main() {
	db, err := gorm.Open(sqlite.Open("math_app.db"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}

	db.AutoMigrate(&Lesson{}, &Task{})

	docsPath := "docs"
	lessonDirs, _ := ioutil.ReadDir(filepath.Join(docsPath, "lesson"))

	lessonTaskMap := make(map[uint][]uint)
	lessonData := make(map[uint]LessonInfo)

	fmt.Println("Scanning lessons...")

	for _, dir := range lessonDirs {
		if !dir.IsDir() {
			continue
		}

		var lessonID int
		fmt.Sscanf(dir.Name(), "%d", &lessonID)
		if lessonID == 0 {
			continue
		}

		indexPath := filepath.Join(docsPath, "lesson", dir.Name(), "index.html")
		content, err := ioutil.ReadFile(indexPath)
		if err != nil {
			continue
		}
		html := string(content)

		titleRe := regexp.MustCompile(`<title>(.+?) \| VisualMath</title>`)
		titleMatch := titleRe.FindStringSubmatch(html)
		title := ""
		if len(titleMatch) > 1 {
			title = titleMatch[1]
		}

		gradeRe := regexp.MustCompile(`/grade/(\d+)`)
		gradeMatch := gradeRe.FindStringSubmatch(html)
		grade := 5
		if len(gradeMatch) > 1 {
			fmt.Sscanf(gradeMatch[1], "%d", &grade)
		}

		canvaRe := regexp.MustCompile(`<iframe[^>]+src="([^"]+canva\.com[^"]+)"`)
		canvaMatch := canvaRe.FindStringSubmatch(html)
		canvaURL := ""
		if len(canvaMatch) > 1 {
			canvaURL = canvaMatch[1]
		}

		descRe := regexp.MustCompile(`<meta name="description" content="([^"]+)"`)
		descMatch := descRe.FindStringSubmatch(html)
		description := ""
		if len(descMatch) > 1 {
			description = descMatch[1]
		}

		taskRe := regexp.MustCompile(`/task/(\d+)`)
		taskMatches := taskRe.FindAllStringSubmatch(html, -1)
		var taskIDs []uint
		for _, m := range taskMatches {
			var taskID uint
			fmt.Sscanf(m[1], "%d", &taskID)
			taskIDs = append(taskIDs, taskID)
		}

		lessonTaskMap[uint(lessonID)] = taskIDs
		lessonData[uint(lessonID)] = LessonInfo{
			Grade:       grade,
			Title:       title,
			CanvaURL:    canvaURL,
			Description: description,
		}

		fmt.Printf("Lesson %d: %s (Grade %d, Tasks: %v)\n", lessonID, title, grade, taskIDs)
	}

	taskDirs, _ := ioutil.ReadDir(filepath.Join(docsPath, "task"))
	taskData := make(map[uint]TaskInfo)

	fmt.Println("\nScanning tasks...")

	for _, dir := range taskDirs {
		if !dir.IsDir() {
			continue
		}

		var taskID int
		fmt.Sscanf(dir.Name(), "%d", &taskID)
		if taskID == 0 {
			continue
		}

		indexPath := filepath.Join(docsPath, "task", dir.Name(), "index.html")
		content, err := ioutil.ReadFile(indexPath)
		if err != nil {
			continue
		}
		html := string(content)

		titleRe := regexp.MustCompile(`<title>(.+?) \| VisualMath</title>`)
		titleMatch := titleRe.FindStringSubmatch(html)
		title := ""
		if len(titleMatch) > 1 {
			title = titleMatch[1]
		}

		descRe := regexp.MustCompile(`<meta name="description" content="([^"]+)"`)
		descMatch := descRe.FindStringSubmatch(html)
		description := ""
		if len(descMatch) > 1 {
			description = descMatch[1]
		}

		srcdocRe := regexp.MustCompile(`srcdoc="([^"]*)"`)
		srcdocMatch := srcdocRe.FindStringSubmatch(html)
		code := ""
		if len(srcdocMatch) > 1 {
			code = srcdocMatch[1]
			code = strings.ReplaceAll(code, "&lt;", "<")
			code = strings.ReplaceAll(code, "&gt;", ">")
			code = strings.ReplaceAll(code, "&amp;", "&")
			code = strings.ReplaceAll(code, "&quot;", "\"")
			code = strings.ReplaceAll(code, "&#34;", "\"")
			code = strings.ReplaceAll(code, "&#39;", "'")
		}

		taskData[uint(taskID)] = TaskInfo{
			Title:       title,
			Description: description,
			Code:        code,
		}

		fmt.Printf("Task %d: %s (Code length: %d)\n", taskID, title, len(code))
	}

	fmt.Println("\n--- Inserting into database ---")

	taskLessonMap := make(map[uint]uint)
	for lessonID, taskIDs := range lessonTaskMap {
		for _, taskID := range taskIDs {
			taskLessonMap[taskID] = lessonID
		}
	}

	for lessonID, info := range lessonData {
		var existing Lesson
		result := db.First(&existing, lessonID)
		if result.Error != nil {
			lesson := Lesson{
				ID:            lessonID,
				Grade:         info.Grade,
				Title:         info.Title,
				CanvaEmbedURL: info.CanvaURL,
				Description:   info.Description,
				SortOrder:     0,
			}
			db.Create(&lesson)
			fmt.Printf("Created Lesson %d: %s\n", lessonID, info.Title)
		} else {
			fmt.Printf("Lesson %d exists, skipping\n", lessonID)
		}
	}

	for taskID, info := range taskData {
		lessonID, found := taskLessonMap[taskID]
		if !found {
			fmt.Printf("Task %d has no lesson, skipping\n", taskID)
			continue
		}

		var existing Task
		result := db.First(&existing, taskID)
		if result.Error != nil {
			task := Task{
				ID:          taskID,
				LessonID:    lessonID,
				Title:       info.Title,
				Description: info.Description,
				Code:        info.Code,
				Order:       1,
			}
			db.Create(&task)
			fmt.Printf("Created Task %d: %s (Lesson %d)\n", taskID, info.Title, lessonID)
		} else {
			fmt.Printf("Task %d exists, skipping\n", taskID)
		}
	}

	fmt.Println("\n✅ Recovery complete!")
}
