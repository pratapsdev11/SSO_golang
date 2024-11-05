package main

import (
    "log"
    "github.com/gin-gonic/gin"
    "your-project/internal/controllers"
    "your-project/internal/middleware"
    "your-project/pkg/database"
    "your-project/pkg/cache"
)

func main() {
    // Initialize database
    db, err := database.InitPostgres()
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // Initialize Redis
    redisClient := cache.InitRedis()

    // Initialize Gin router
    r := gin.Default()

    // CORS middleware
    r.Use(middleware.CorsMiddleware())

    // Initialize controllers
    authController := controllers.NewAuthController(db, redisClient)
    sessionController := controllers.NewSessionController(db, redisClient)

    // API routes
    api := r.Group("/api")
    {
        auth := api.Group("/auth")
        {
            auth.POST("/microsoft", authController.MicrosoftAuth)
            auth.POST("/validate", authController.ValidateToken)
        }

        session := api.Group("/session")
        session.Use(middleware.AuthMiddleware())
        {
            session.POST("/create", sessionController.Create)
            session.DELETE("/:id", sessionController.Delete)
        }
    }

    r.Run(":8080")
}