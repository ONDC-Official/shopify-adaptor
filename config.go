package main

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

// initConfig loads the .env file (if available) and sets default values.
func initConfig() {
    // Load .env file if available.
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using system environment variables")
    }

    // Set default environment variables if not already set.
    if os.Getenv("SHOPIFY_URL") == "" {
        os.Setenv("SHOPIFY_URL", "https://testgamaa.myshopify.com")
    }

    log.Printf("Shopify URL: %s", os.Getenv("SHOPIFY_URL"))
    log.Printf("Access Token configured: %v", os.Getenv("SHOPIFY_ACCESS_TOKEN") != "")
}
