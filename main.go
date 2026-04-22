package main

import (
	"log"
	"os"

	"github.com/Mobilizes/materi-be-alpro/config"
	"github.com/Mobilizes/materi-be-alpro/database/entities"
	"github.com/Mobilizes/materi-be-alpro/database/seeders"
	"github.com/Mobilizes/materi-be-alpro/modules/user"
	userController "github.com/Mobilizes/materi-be-alpro/modules/user/controller"
	userRepository "github.com/Mobilizes/materi-be-alpro/modules/user/repository"
	userService "github.com/Mobilizes/materi-be-alpro/modules/user/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/Mobilizes/materi-be-alpro/modules/auth"
	authController "github.com/Mobilizes/materi-be-alpro/modules/auth/controller"
	authService "github.com/Mobilizes/materi-be-alpro/modules/auth/service"
)

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8880"
	}

	// Connect to database
	db := config.SetupDatabase()

	// Auto migrate the user module
	db.AutoMigrate(&entities.User{})
	seeders.SeedUsers(db)

	// Initialize Gin app
	r := gin.Default()

	// Setup routes group
	api := r.Group("/api")

	// Setup Services
	jwtSvc := authService.NewJWTService()
	userRepo := userRepository.NewUserRepository(db)
	uService := userService.NewUserService(userRepo)
	aService := authService.NewAuthService(userRepo, jwtSvc)

	// Setup Controllers
	userCtrl := userController.NewUserController(uService)
	authCtrl := authController.NewAuthController(aService)

	// Register Routes
	auth.RegisterAuthRoutes(api, authCtrl)
	user.RegisterUserRoutes(api, userCtrl, jwtSvc)

	// Start App
	r.Run(":" + port)
}
