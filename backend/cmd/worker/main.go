package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/digi604/swarmmarket/backend/internal/auction"
	"github.com/digi604/swarmmarket/backend/internal/config"
	"github.com/digi604/swarmmarket/backend/internal/database"
	"github.com/digi604/swarmmarket/backend/internal/email"
	"github.com/digi604/swarmmarket/backend/internal/notification"
	"github.com/digi604/swarmmarket/backend/internal/worker"
)

func main() {
	// Load configuration
	cfg := config.MustLoad()
	log.Println("Worker: Configuration loaded")

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Connect to PostgreSQL
	initCtx, initCancel := context.WithTimeout(ctx, 30*time.Second)
	db, err := database.NewPostgresDB(initCtx, cfg.Database)
	initCancel()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Worker: Connected to PostgreSQL")

	// Connect to Redis
	initCtx, initCancel = context.WithTimeout(ctx, 30*time.Second)
	redis, err := database.NewRedisDB(initCtx, cfg.Redis)
	initCancel()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()
	log.Println("Worker: Connected to Redis")

	// Initialize services
	notificationService := notification.NewService(redis.Client)
	webhookRepo := notification.NewRepository(db.Pool)
	auctionRepo := auction.NewRepository(db.Pool)
	auctionService := auction.NewService(auctionRepo, notificationService)

	// Initialize email service (SendGrid)
	var emailService *email.Service
	if cfg.Email.SendGridAPIKey != "" {
		emailRepo := email.NewRepository(db.Pool)
		emailService = email.NewService(
			cfg.Email.SendGridAPIKey,
			cfg.Email.FromEmail,
			cfg.Email.FromName,
			cfg.Server.PublicURL,
			cfg.Email.CooldownMinutes,
			emailRepo,
		)
		log.Println("Worker: Email service initialized (SendGrid)")
	} else {
		log.Println("Worker: SendGrid not configured - email notifications disabled")
	}

	// Create worker
	w := worker.New(worker.Config{
		NotificationService: notificationService,
		WebhookRepo:         webhookRepo,
		AuctionService:      auctionService,
		AuctionRepo:         auctionRepo,
		EmailService:        emailService,
		RedisClient:         redis.Client,
	})

	// Start worker
	go func() {
		if err := w.Run(ctx); err != nil {
			log.Printf("Worker error: %v", err)
		}
	}()
	log.Println("Worker: Started")

	// Wait for shutdown signal
	<-sigChan
	log.Println("Worker: Shutting down...")
	cancel()

	// Give workers time to finish
	time.Sleep(2 * time.Second)
	log.Println("Worker: Stopped")
}
