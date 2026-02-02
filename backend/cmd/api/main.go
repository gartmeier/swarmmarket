package main

import (
	"context"
	"log"
	"time"

	"github.com/digi604/swarmmarket/backend/internal/agent"
	"github.com/digi604/swarmmarket/backend/internal/auction"
	"github.com/digi604/swarmmarket/backend/internal/capability"
	"github.com/digi604/swarmmarket/backend/internal/config"
	"github.com/digi604/swarmmarket/backend/internal/database"
	"github.com/digi604/swarmmarket/backend/internal/marketplace"
	"github.com/digi604/swarmmarket/backend/internal/matching"
	"github.com/digi604/swarmmarket/backend/internal/notification"
	"github.com/digi604/swarmmarket/backend/internal/payment"
	"github.com/digi604/swarmmarket/backend/internal/transaction"
	"github.com/digi604/swarmmarket/backend/internal/user"
	"github.com/digi604/swarmmarket/backend/internal/wallet"
	"github.com/digi604/swarmmarket/backend/pkg/api"
	"github.com/digi604/swarmmarket/backend/pkg/websocket"
)

func main() {
	// Load configuration
	cfg := config.MustLoad()
	log.Println("Configuration loaded")

	// Create context with timeout for initialization
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to PostgreSQL
	db, err := database.NewPostgresDB(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL")

	// Run migrations
	if err := database.RunMigrations(ctx, db.Pool); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed")

	// Connect to Redis
	redis, err := database.NewRedisDB(ctx, cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()
	log.Println("Connected to Redis")

	// Initialize services
	agentRepo := agent.NewRepository(db.Pool)
	agentService := agent.NewService(agentRepo, cfg.Auth.APIKeyLength)

	// Initialize notification service (implements EventPublisher for marketplace)
	notificationService := notification.NewService(redis.Client)

	// Initialize marketplace service
	marketplaceRepo := marketplace.NewRepository(db.Pool)
	marketplaceService := marketplace.NewService(marketplaceRepo, notificationService)

	// Initialize capability service
	capabilityRepo := capability.NewRepository(db.Pool)
	capabilityService := capability.NewService(capabilityRepo)

	// Initialize transaction service
	transactionRepo := transaction.NewRepository(db.Pool)
	transactionService := transaction.NewService(transactionRepo, notificationService)

	// Wire transaction creator to marketplace service (avoids circular dependency)
	marketplaceService.SetTransactionCreator(transactionService)

	// Initialize auction service
	auctionRepo := auction.NewRepository(db.Pool)
	auctionService := auction.NewService(auctionRepo, notificationService)

	// Initialize webhook repository (for notification management)
	webhookRepo := notification.NewRepository(db.Pool)

	// Initialize WebSocket hub for real-time notifications
	wsHub := websocket.NewHub()
	go wsHub.Run(context.Background())
	log.Println("WebSocket hub started")

	// Initialize matching engine (NYSE-style order book)
	matchingEngine := matching.NewEngine(func(ctx context.Context, trade matching.Trade) {
		// Publish trade events
		notificationService.Publish(ctx, "match.found", map[string]any{
			"trade_id":   trade.ID,
			"product_id": trade.ProductID,
			"buyer_id":   trade.BuyerID,
			"seller_id":  trade.SellerID,
			"price":      trade.Price,
			"quantity":   trade.Quantity,
		})
	})
	log.Println("Matching engine initialized")

	// Initialize payment service (Stripe)
	var paymentService *payment.Service
	if cfg.Stripe.SecretKey != "" {
		paymentService = payment.NewService(payment.Config{
			SecretKey:          cfg.Stripe.SecretKey,
			WebhookSecret:      cfg.Stripe.WebhookSecret,
			PlatformFeePercent: cfg.Stripe.PlatformFeePercent,
			DefaultReturnURL:   cfg.Stripe.DefaultReturnURL,
		})
		// Wire payment adapter to transaction service for escrow
		transactionService.SetPaymentService(payment.NewAdapter(paymentService))
		log.Println("Stripe payment service initialized and wired to transactions")
	} else {
		log.Println("Stripe not configured - payment endpoints disabled")
	}

	// Initialize user service (for human dashboard)
	var userService *user.Service
	if cfg.Clerk.SecretKey != "" {
		userRepo := user.NewRepository(db.Pool)
		userService = user.NewService(userRepo)
		log.Println("User service initialized (Clerk authentication enabled)")
	} else {
		log.Println("Clerk not configured - dashboard endpoints disabled")
	}

	// Initialize wallet service (for deposits)
	var walletService *wallet.Service
	if cfg.Stripe.SecretKey != "" {
		walletRepo := wallet.NewRepository(db.Pool)
		walletService = wallet.NewService(walletRepo, wallet.StripeConfig{
			SecretKey: cfg.Stripe.SecretKey,
		})
		log.Println("Wallet service initialized")

		// Wire wallet balance checker to marketplace and auction services
		// This enforces that agents have sufficient funds before accepting offers or placing bids
		balanceChecker := wallet.NewBalanceChecker(walletService)
		marketplaceService.SetWalletChecker(balanceChecker)
		auctionService.SetWalletChecker(balanceChecker)
		log.Println("Wallet balance checker wired to marketplace and auction services")
	}

	// Create router
	router := api.NewRouter(api.RouterConfig{
		Config:             cfg,
		AgentService:       agentService,
		MarketplaceService: marketplaceService,
		CapabilityService:  capabilityService,
		TransactionService: transactionService,
		AuctionService:     auctionService,
		MatchingEngine:     matchingEngine,
		PaymentService:     paymentService,
		WalletService:      walletService,
		WebhookRepo:        webhookRepo,
		WebSocketHub:       wsHub,
		UserService:        userService,
		DB:                 db,
		Redis:              redis,
	})

	// Create and run server
	server := api.NewServer(cfg.Server, router)
	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped")
}
