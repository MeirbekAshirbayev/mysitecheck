package handlers

import (
	"fmt"
	"math-app/internal/builder"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// AdminExportHandler builds the static site
func AdminExportHandler(c *gin.Context) {
	cwd, err := os.Getwd()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to get CWD: %v", err)
		return
	}
	// Save to "site_export" to be very clear
	distDir := filepath.Join(cwd, "docs")

	if err := builder.BuildSite(distDir, "/mysitecheck"); err != nil {
		c.String(http.StatusInternalServerError, "Export Failed: %v", err)
		return
	}

	// Simple success page
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, fmt.Sprintf(`
        <html>
        <head>
			<title>Export Success</title>
			<script src="https://cdn.tailwindcss.com"></script>
		</head>
        <body class="bg-gray-100 flex items-center justify-center h-screen font-sans">
            <div class="bg-white p-8 rounded-xl shadow-lg text-center max-w-lg">
                <div class="text-6xl mb-4">✅</div>
                <h1 class="text-2xl font-bold text-green-600 mb-4">Сәтті Экспортталды!</h1>
                <p class="mb-6 text-gray-700">Сайттың барлық HTML файлдары мына папкаға сақталды:</p>
				<code class="block bg-gray-800 text-yellow-300 p-3 rounded text-sm mb-6 break-all">%s</code>
                <div class="space-x-4">
                     <a href="http://localhost:8081" target="_blank" class="bg-green-600 text-white px-6 py-2 rounded-lg hover:bg-green-700 transition">👁️ Тексеру (Preview)</a>
                     <a href="/admin/" class="bg-indigo-600 text-white px-6 py-2 rounded-lg hover:bg-indigo-700 transition">Админге оралу</a>
                </div>
            </div>
        </body>
        </html>
    `, distDir))
}
