package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"proclients/backend/internal/auth"
	"proclients/backend/internal/config"
	"proclients/backend/internal/handler"
	"proclients/backend/internal/repository"
	"proclients/backend/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connection error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("db ping error: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	leadRepo := repository.NewLeadRepository(db)
	dealRepo := repository.NewDealRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	catalogProductRepo := repository.NewCatalogProductRepository(db)
	attachmentRepo := repository.NewAttachmentRepository(db)
	activityRepo := repository.NewActivityRepository(db)

	jwtManager, err := auth.NewManager(cfg.JWTSecret, cfg.JWTTTL)
	if err != nil {
		log.Fatalf("jwt config error: %v", err)
	}

	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo)
	activityService := service.NewActivityService(activityRepo)
	attachmentService := service.NewAttachmentService(attachmentRepo, activityService)
	leadService := service.NewLeadService(leadRepo, attachmentService, activityService)
	dealService := service.NewDealService(dealRepo, attachmentService, activityService)
	taskService := service.NewTaskService(taskRepo, attachmentService, activityService)
	catalogProductService := service.NewCatalogProductService(catalogProductRepo)

	beelineIntegrationService := service.NewBeelineIntegrationService(
		leadService,
		leadRepo,
		cfg.BeelineWebhookSecret,
		cfg.BeelineCreatedByUser,
	)

	authHandler := handler.NewAuthHandler(authService, jwtManager)
	leadHandler := handler.NewLeadHandler(leadService, attachmentService)
	dealHandler := handler.NewDealHandler(dealService, attachmentService)
	taskHandler := handler.NewTaskHandler(taskService, attachmentService)
	catalogProductHandler := handler.NewCatalogProductHandler(catalogProductService)
	userHandler := handler.NewUserHandler(userService)
	attachmentHandler := handler.NewAttachmentHandler(attachmentService)
	beelineHandler := handler.NewBeelineIntegrationHandler(beelineIntegrationService)

	router := handler.NewRouter(
		authHandler,
		leadHandler,
		dealHandler,
		taskHandler,
		catalogProductHandler,
		userHandler,
		attachmentHandler,
		beelineHandler,
		jwtManager,
		cfg.CORSOrigins,
	)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("server listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
