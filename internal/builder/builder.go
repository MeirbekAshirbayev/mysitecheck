package builder

import (
	"fmt"
	"html/template"
	"math-app/internal/database"
	"math-app/internal/models"
	"os"
	"path/filepath"
	"strings"
)

// RenderToFile renders a template to a file
func RenderToFile(path string, tmplName string, data interface{}, funcMap template.FuncMap) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Parse templates (Doing this every time is slightly inefficient but safe and simple)
	tmpl, err := template.New(tmplName).Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := tmpl.ExecuteTemplate(f, tmplName, data); err != nil {
		return err
	}
	return nil
}

// BuildSite generates the static site in the outputDir
// Set basePath to "/mysitecheck" for GitHub Pages or "/" for root deployment
func BuildSite(outputDir string, basePath string) error {
	// 1. Clean/Create Dir
	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("failed to clear output dir: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %v", err)
	}

	// Ensure basePath ends without trailing slash for consistency
	if basePath != "/" && len(basePath) > 0 && basePath[len(basePath)-1] == '/' {
		basePath = basePath[:len(basePath)-1]
	}

	funcMap := template.FuncMap{
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
		"path": func(p string) string {
			p = strings.TrimSpace(p) // Remove leading/trailing spaces
			if basePath == "/" {
				return p
			}
			return basePath + p
		},
	}

	fmt.Println("Building Index...")
	// 2. Index Page
	var grades []int
	database.DB.Model(&models.Lesson{}).Distinct("grade").Order("grade").Pluck("grade", &grades)
	if err := RenderToFile(filepath.Join(outputDir, "index.html"), "index.html", map[string]interface{}{"Grades": grades}, funcMap); err != nil {
		return fmt.Errorf("failed to render index: %v", err)
	}

	// 3. Grade Pages
	fmt.Println("Building Grade Pages...")
	for _, g := range grades {
		var lessons []models.Lesson
		database.DB.Where("grade = ?", g).Find(&lessons)
		path := filepath.Join(outputDir, fmt.Sprintf("grade/%d/index.html", g))
		if err := RenderToFile(path, "grade_list.html", map[string]interface{}{"Grade": g, "Lessons": lessons}, funcMap); err != nil {
			return fmt.Errorf("failed to render grade %d: %v", g, err)
		}
	}

	// 4. Lessons
	fmt.Println("Building Lessons...")
	var allLessons []models.Lesson
	database.DB.Preload("Tasks").Find(&allLessons)
	for _, l := range allLessons {
		path := filepath.Join(outputDir, fmt.Sprintf("lesson/%d/index.html", l.ID))
		if err := RenderToFile(path, "lesson.html", map[string]interface{}{"Lesson": l}, funcMap); err != nil {
			return fmt.Errorf("failed to render lesson %d: %v", l.ID, err)
		}
	}

	// 5. Tasks
	fmt.Println("Building Tasks...")
	var allTasks []models.Task
	database.DB.Find(&allTasks)
	for _, t := range allTasks {
		path := filepath.Join(outputDir, fmt.Sprintf("task/%d/index.html", t.ID))
		if err := RenderToFile(path, "task.html", map[string]interface{}{"Task": t}, funcMap); err != nil {
			return fmt.Errorf("failed to render task %d: %v", t.ID, err)
		}
	}

	// 6. Static Pages
	fmt.Println("Building Static Pages...")
	if err := RenderToFile(filepath.Join(outputDir, "privacy/index.html"), "privacy.html", nil, funcMap); err != nil {
		return err
	}
	if err := RenderToFile(filepath.Join(outputDir, "terms/index.html"), "terms.html", nil, funcMap); err != nil {
		return err
	}

	fmt.Println("Build Complete!")
	return nil
}
