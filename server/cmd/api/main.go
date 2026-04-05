package main

import (
	"cmd/api/internal/auth"
	"cmd/api/internal/config"
	"cmd/api/internal/database"
	"cmd/api/internal/handlers"
	"cmd/api/internal/middleware"
	"cmd/api/internal/repository"
	"cmd/api/internal/routes"
	"cmd/api/internal/service"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiration)

	userRepo := repository.NewUserRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	lineItemRepo := repository.NewLineItemRepository(db)
	voucherRepo := repository.NewVoucherRepository(db)
	reportRepo := repository.NewReportRepository(db)
	scheduledTaskRepo := repository.NewScheduledTaskRepository(db)
	agentMessageRepo := repository.NewAgentMessageRepository(db)
	agentUsageRepo := repository.NewAgentUsageRepository(db)

	userService := service.NewUserService(userRepo)
	accountService := service.NewAccountService(accountRepo)
	lineItemService := service.NewLineItemService(lineItemRepo)
	voucherService := service.NewVoucherService(voucherRepo, lineItemRepo)
	reportService := service.NewReportService(reportRepo)
	scheduledTaskService := service.NewScheduledTaskService(scheduledTaskRepo)
	agentService := service.NewAgentService(
		cfg.GeminiAPIKey,
		voucherService,
		accountService,
		lineItemService,
		reportService,
		scheduledTaskService,
		agentUsageRepo,
	)

	userHandler := handlers.NewUserHandler(userService)
	accountHandler := handlers.NewAccountHandler(accountService)
	lineItemHandler := handlers.NewLineItemHandler(lineItemService)
	voucherHandler := handlers.NewVoucherHandler(voucherService)
	authHandler := handlers.NewAuthHandler(userService, jwtManager)
	pdfHandler := handlers.NewPDFHandler(voucherService, accountService)
	reportHandler := handlers.NewReportHandler(reportService)
	sieGenerator := service.NewSIEGenerator(reportRepo, userService)
	sieHandler := handlers.NewSIEHandler(sieGenerator)
	agentHandler := handlers.NewAgentHandler(agentService, scheduledTaskService, agentMessageRepo, agentUsageRepo)

	authMiddleware := middleware.AuthMiddleware(jwtManager)

	router := gin.Default()

	// Add CORS middleware
	router.Use(middleware.CORSMiddleware())

	routes.SetupRoutes(router, userHandler, accountHandler, lineItemHandler, voucherHandler, authHandler, pdfHandler, reportHandler, sieHandler, agentHandler, authMiddleware)

	// Start the scheduler for recurring tasks
	schedulerService := service.NewSchedulerService(scheduledTaskService, agentService)
	schedulerService.Start()
	defer schedulerService.Stop()

	log.Println("Starting server on", cfg.ServerPort)
	if err := router.Run(cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}