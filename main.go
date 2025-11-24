package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	} else {
		log.Println("✓ .env file loaded successfully")
	}

	// Verify critical environment variables
	region := os.Getenv("AWS_REGION")
	bucket := os.Getenv("S3_BUCKET")

	log.Printf("AWS_REGION: %s", region)
	log.Printf("S3_BUCKET: %s", bucket)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// load and validate config (loads CUSTOMER_KEY_BASE64 into runtime key)
	if err := loadRuntimeConfig(); err != nil {
		log.Fatalf("config: %v", err)
	}

	r := newRouter()
	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
