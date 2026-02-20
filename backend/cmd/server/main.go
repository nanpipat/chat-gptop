package main

import (
	"context"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"rag-chat-system/internal/config"
	"rag-chat-system/internal/database"
	"rag-chat-system/internal/handlers"
	"rag-chat-system/internal/repositories"
	"rag-chat-system/internal/services"
	"rag-chat-system/internal/storage"
)

func main() {
	cfg := config.Load()

	// Database
	pool := database.Connect(cfg.DatabaseURL())
	database.Migrate(pool)

	// Repositories
	projectRepo := repositories.NewProjectRepo(pool)
	fileRepo := repositories.NewFileRepo(pool)
	chunkRepo := repositories.NewChunkRepo(pool)
	chatRepo := repositories.NewChatRepo(pool)
	messageRepo := repositories.NewMessageRepo(pool)

	// Storage
	var store storage.Storage
	if cfg.R2Endpoint != "" && cfg.R2AccessKeyID != "" && cfg.R2SecretAccessKey != "" && cfg.R2Bucket != "" {
		log.Println("Initializing S3 Storage (R2)...")
		s3Store, err := storage.NewS3Storage(cfg.R2AccessKeyID, cfg.R2SecretAccessKey, cfg.R2Endpoint, cfg.R2Bucket)
		if err != nil {
			log.Fatalf("Failed to initialize S3 storage: %v", err)
		}

		// Verify connection immediately
		log.Println("Verifying R2 connection...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := s3Store.VerifyConnection(ctx); err != nil {
			cancel()
			log.Fatalf("❌ R2 Connection Verification Failed: %v\nCheck R2_ENDPOINT, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, and R2_BUCKET", err)
		}
		cancel()
		log.Println("✅ R2 Connection Verified Successfully")

		store = s3Store
	} else {
		log.Println("Initializing Local Storage...")
		store = storage.NewLocalStorage(cfg.StoragePath)
	}

	// Services
	openaiSvc := services.NewOpenAIService(cfg.OpenAIKey)
	embeddingSvc := services.NewEmbeddingService(openaiSvc)
	ragSvc := services.NewRAGService(chunkRepo, embeddingSvc)
	ingestSvc := services.NewIngestService(chunkRepo, embeddingSvc)
	fileSvc := services.NewFileService(fileRepo, chunkRepo, ingestSvc, store)
	chatSvc := services.NewChatService(chatRepo, messageRepo, ragSvc, openaiSvc)
	gitSvc := services.NewGitService(projectRepo, fileRepo, chunkRepo, fileSvc, cfg.GitEncryptionKey)

	// Handlers
	projectHandler := handlers.NewProjectHandler(projectRepo)
	fileHandler := handlers.NewFileHandler(fileSvc)
	chatHandler := handlers.NewChatHandler(chatSvc)
	gitHandler := handlers.NewGitHandler(gitSvc)

	// Echo
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"*"},
	}))
	e.Use(middleware.Logger())

	// Projects
	e.POST("/projects", projectHandler.Create)
	e.GET("/projects", projectHandler.List)
	e.DELETE("/projects/:id", projectHandler.Delete)

	// Files
	e.POST("/projects/:id/upload-file", fileHandler.UploadFile)
	e.POST("/projects/:id/upload-folder", fileHandler.UploadFolder)
	e.GET("/projects/:id/files", fileHandler.ListFiles)
	e.DELETE("/files/:id", fileHandler.DeleteFile)

	// Git
	e.PUT("/projects/:id/git", gitHandler.SaveGitConfig)
	e.GET("/projects/:id/git", gitHandler.GetGitConfig)
	e.POST("/projects/:id/git/sync", gitHandler.SyncGit)
	e.DELETE("/projects/:id/git", gitHandler.RemoveGitConfig)

	// Chats
	e.POST("/chats", chatHandler.CreateChat)
	e.GET("/chats", chatHandler.ListChats)
	e.DELETE("/chats/:id", chatHandler.DeleteChat)
	e.PUT("/chats/:id/projects", chatHandler.UpdateChatProjects)
	e.GET("/chats/:id/messages", chatHandler.GetMessages)
	e.POST("/chats/:id/messages", chatHandler.SendMessage)

	log.Fatal(e.Start(":" + cfg.BackendPort))
}
