package main

import (
	"fmt"
	"log"
	"math-app/internal/builder"
	"math-app/internal/database"
	"os"
)

func main() {
	// Initialize DB (needed for builder)
	database.InitDB()

	cwd, _ := os.Getwd()
	fmt.Printf("Building site to 'docs' in %s\n", cwd)

	if err := builder.BuildSite("docs", "/"); err != nil {
		log.Fatalf("Build failed: %v", err)
	}
	fmt.Println("Build success!")
}
