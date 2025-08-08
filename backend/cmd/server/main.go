package main

import (
	"actuworry/backend/handlers"
	"actuworry/backend/routes"
	"actuworry/backend/services"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Initialize service
	actuarialService := services.NewActuarialService()
	
	// Load mortality tables
	tables := []string{"male", "female"}
	for _, tableName := range tables {
		filePath := fmt.Sprintf("backend/data/%s.csv", tableName)
		if err := actuarialService.LoadMortalityTable(tableName, filePath); err != nil {
			log.Fatalf("Failed to load mortality table %s: %v", tableName, err)
		}
		log.Printf("Successfully loaded mortality table: %s", tableName)
	}
	
	// Initialize handlers
	actuarialHandler := handlers.NewActuarialHandler(actuarialService)
	
	// Setup routes
	mux := routes.SetupRoutes(actuarialHandler)
	
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	// Start server
	serverAddr := fmt.Sprintf(":%s", port)
	fmt.Printf("\n🚀 Actuworry Server starting on port %s\n", port)
	fmt.Printf("📊 API Documentation: http://localhost:%s/api/health\n", port)
	fmt.Printf("🌐 Frontend: http://localhost:%s\n", port)
	fmt.Println("\n✅ Server is ready to accept requests")
	
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
