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
	"blog-platform-backend/internal/search"
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
		&models.Follow{},
		&models.Category{},
		&models.Post{},
	); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}
	log.Println("Database migration completed")

	// Migrate single category_id to many-to-many
	migratePostCategories(db)

	// Backfill slugs for existing creators
	backfillSlugs(db)

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

	// Initialize search service
	searchSvc := search.NewService(cfg.MeiliURL, cfg.MeiliMasterKey)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	postHandler := handlers.NewPostHandler(db, rdb, searchSvc)
	categoryHandler := handlers.NewCategoryHandler(db, rdb)
	uploadHandler := handlers.NewUploadHandler(minioClient, cfg)
		channelHandler := handlers.NewChannelHandler(db, cfg)
	searchHandler := handlers.NewSearchHandler(searchSvc)

	// Sync existing data to Meilisearch (non-blocking)
	go func() {
		log.Println("Meilisearch: syncing existing data...")
		searchSvc.SyncAllPosts(db)
		searchSvc.SyncAllCreators(db)
		log.Println("Meilisearch: initial sync complete")
	}()

	// Routes
	api := router.Group("/api/v1")
	{
		// Health check (no auth)
		api.GET("/health", handlers.HealthCheck)

		// Public routes
		api.GET("/posts", postHandler.ListPosts)
		api.GET("/posts/:slug", postHandler.GetPost)
		api.GET("/categories", categoryHandler.ListCategories)
		api.GET("/files/:filename", uploadHandler.ServeFile)
		api.GET("/search", searchHandler.Search)

			// Channel routes
			api.GET("/channels/:slug", channelHandler.GetChannel)
			api.GET("/channels/:slug/is-following", middleware.AuthMiddleware(cfg.JWTSecret), channelHandler.IsFollowing)
			api.POST("/channels/:slug/follow", middleware.AuthMiddleware(cfg.JWTSecret), channelHandler.Follow)
			api.DELETE("/channels/:slug/follow", middleware.AuthMiddleware(cfg.JWTSecret), channelHandler.Unfollow)

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.AuthMiddleware(cfg.JWTSecret), authHandler.Me)
			auth.GET("/google/login", authHandler.GoogleLogin)
			auth.GET("/google/callback", authHandler.GoogleCallback)
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
				creator.GET("/stats", channelHandler.StudioStats)
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
			Slug:         "test-creator",
			DisplayName:  "Test Creator",
		}
		if err := db.Create(&creator).Error; err != nil {
			log.Printf("Warning: failed to seed test creator: %v", err)
		} else {
			log.Printf("Seeded test creator: email=%s, password=Test123!", creator.Email)
		}
	}

	// Seed categories
	categoryNames := []string{"Technology", "Design", "Tutorial", "News", "Lifestyle", "Programming", "DevOps", "Cloud", "AI", "Security", "Database", "Frontend", "Backend", "Mobile", "Game Dev"}
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

// migratePostCategories migrates existing posts with category_id to the many-to-many
// post_categories join table. This is a one-time migration for existing data.
func migratePostCategories(db *gorm.DB) {
	// Check if the old category_id column exists (it will have been dropped by AutoMigrate
	// if the model no longer has it, but we can detect this by looking at the join table)
	if !db.Migrator().HasColumn(&models.Post{}, "category_id") {
		return
	}

	type OldPost struct {
		ID         string
		CategoryID *int
	}
	var oldPosts []OldPost
	db.Model(&models.Post{}).Where("category_id IS NOT NULL").Find(&oldPosts)
	if len(oldPosts) == 0 {
		return
	}

	log.Printf("Migrating %d posts from single category to many-to-many...", len(oldPosts))
	for _, p := range oldPosts {
		if p.CategoryID != nil {
			if err := db.Exec(
				"INSERT INTO post_categories (post_id, category_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
				p.ID, *p.CategoryID,
			).Error; err != nil {
				log.Printf("Warning: failed to migrate post %s category %d: %v", p.ID, *p.CategoryID, err)
			}
		}
	}
	log.Printf("Category migration complete")
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
// backfillSlugs generates slugs for any creators that don't have one yet.
// This handles existing records after the slug column was added.
func backfillSlugs(db *gorm.DB) {
	var creators []models.Creator
	db.Where("slug IS NULL OR slug = ''").Find(&creators)
	for _, c := range creators {
		slug := models.GenerateSlug(c.DisplayName)
		// Check uniqueness and append suffix if needed
		for {
			var count int64
			db.Model(&models.Creator{}).Where("slug = ?", slug).Count(&count)
			if count == 0 {
				break
			}
			slug = models.AppendRandomSuffix(slug)
		}
		db.Model(&c).Update("slug", slug)
		log.Printf("Backfilled slug for creator %s: %s", c.Email, slug)
	}
}

