package main

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"blog-platform-backend/internal/config"
	"blog-platform-backend/internal/database"
	"blog-platform-backend/internal/handlers"
	"blog-platform-backend/internal/middleware"
	"blog-platform-backend/internal/models"
	"blog-platform-backend/internal/storage"
)

func main() {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	// Load configuration
	cfg := config.Load()

	// Connect to PostgreSQL
	db := database.Connect(cfg.DatabaseURL)

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.Creator{},
		&models.Category{},
		&models.Post{},
	); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}
	log.Println("Database migration completed")

	// Seed initial data
	seedData(db)

	// Connect to Redis
	rdb := storage.NewRedisClient(cfg.RedisURL)

	// Connect to MinIO
	minioClient := storage.NewMinIOClient(
		cfg.MinIOEndpoint,
		cfg.MinIOAccessKey,
		cfg.MinIOSecretKey,
		cfg.MinIOBucket,
		cfg.MinIOUseSSL,
	)

	// Setup Gin router
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	postHandler := handlers.NewPostHandler(db, rdb)
	categoryHandler := handlers.NewCategoryHandler(db, rdb)
	uploadHandler := handlers.NewUploadHandler(minioClient, cfg)

	// Routes
	api := router.Group("/api/v1")
	{
		// Public routes
		api.GET("/posts", postHandler.ListPosts)
		api.GET("/posts/:slug", postHandler.GetPost)
		api.GET("/categories", categoryHandler.ListCategories)
		api.GET("/files/:filename", uploadHandler.ServeFile)

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.AuthMiddleware(cfg.JWTSecret), authHandler.Me)
		}

		// Creator routes (auth required)
		creator := api.Group("/creator")
		creator.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			creator.GET("/posts", postHandler.CreatorListPosts)
			creator.POST("/posts", postHandler.CreatorCreatePost)
			creator.PUT("/posts/:id", postHandler.CreatorUpdatePost)
			creator.DELETE("/posts/:id", postHandler.CreatorDeletePost)
			creator.POST("/upload", uploadHandler.Upload)
		}
	}

	// Start server
	addr := ":" + cfg.ServerPort
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// seedData populates the database with initial test data if empty.
func seedData(db *gorm.DB) {
	// Seed test creator
	var creatorCount int64
	db.Model(&models.Creator{}).Count(&creatorCount)
	if creatorCount == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Test123!"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		creator := models.Creator{
			Email:        "creator@test.com",
			PasswordHash: string(hashedPassword),
			DisplayName:  "Test Creator",
		}
		if err := db.Create(&creator).Error; err != nil {
			log.Printf("Warning: failed to seed test creator: %v", err)
		} else {
			log.Printf("Seeded test creator: email=%s, password=Test123!", creator.Email)
		}
	}

	// Seed categories
	categoryNames := []string{"Technology", "Design", "Tutorial", "News", "Lifestyle"}
	for _, name := range categoryNames {
		var count int64
		db.Model(&models.Category{}).Where("name = ?", name).Count(&count)
		if count == 0 {
			category := models.Category{
				Name: name,
				Slug: slugify(name),
			}
			if err := db.Create(&category).Error; err != nil {
				log.Printf("Warning: failed to seed category %q: %v", name, err)
			} else {
				log.Printf("Seeded category: %s", name)
			}
		}
	}
}

// slugify creates a URL-safe slug from a string.
func slugify(s string) string {
	slug := strings.ToLower(s)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove non-alphanumeric, non-hyphen characters
	cleaned := make([]byte, 0, len(slug))
	for i := 0; i < len(slug); i++ {
		c := slug[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			cleaned = append(cleaned, c)
		}
	}
	slug = string(cleaned)
	// Collapse consecutive hyphens
	result := make([]byte, 0, len(slug))
	lastHyphen := false
	for i := 0; i < len(slug); i++ {
		c := slug[i]
		if c == '-' {
			if lastHyphen {
				continue
			}
			lastHyphen = true
		} else {
			lastHyphen = false
		}
		result = append(result, c)
	}
	slug = string(result)
	// Trim leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	return slug
}