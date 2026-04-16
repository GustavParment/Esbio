package main

import (
	"esbio/api/internal/auth"
	"esbio/api/internal/config"
	"esbio/api/internal/database"
	"esbio/api/internal/handlers"
	"esbio/api/internal/middleware"
	"esbio/api/internal/repository"
	"esbio/api/internal/routes"
	"esbio/api/internal/service"
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
	companyRepo := repository.NewCompanyRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	lineItemRepo := repository.NewLineItemRepository(db)
	voucherRepo := repository.NewVoucherRepository(db)
	reportRepo := repository.NewReportRepository(db)
	scheduledTaskRepo := repository.NewScheduledTaskRepository(db)
	agentMessageRepo := repository.NewAgentMessageRepository(db)
	agentUsageRepo := repository.NewAgentUsageRepository(db)

	userService := service.NewUserService(userRepo)
	companyService := service.NewCompanyService(companyRepo)
	accountService := service.NewAccountService(accountRepo)
	lineItemService := service.NewLineItemService(lineItemRepo)
	voucherService := service.NewVoucherService(voucherRepo, lineItemRepo)
	reportService := service.NewReportService(reportRepo)
	scheduledTaskService := service.NewScheduledTaskService(scheduledTaskRepo)
	receiptService := service.NewReceiptService(cfg.GeminiAPIKey)
	stripeService := service.NewStripeService(cfg.StripeSecretKey, cfg.StripeWebhookSecret, cfg.FrontendURL, companyRepo)
	agentService := service.NewAgentService(
		cfg.GeminiAPIKey,
		voucherService,
		accountService,
		lineItemService,
		reportService,
		scheduledTaskService,
		agentUsageRepo,
	)

	emailService := service.NewEmailService(cfg.ResendAPIKey)

	userHandler := handlers.NewUserHandler(userService)
	accountHandler := handlers.NewAccountHandler(accountService)
	lineItemHandler := handlers.NewLineItemHandler(lineItemService)
	voucherHandler := handlers.NewVoucherHandler(voucherService)
	authHandler := handlers.NewAuthHandler(userService, jwtManager, emailService, userRepo, cfg.FrontendURL)
	pdfHandler := handlers.NewPDFHandler(voucherService, accountService)
	reportHandler := handlers.NewReportHandler(reportService)
	companyHandler := handlers.NewCompanyHandler(companyService)
	sieGenerator := service.NewSIEGenerator(reportRepo, companyService)
	sieHandler := handlers.NewSIEHandler(sieGenerator)
	receiptHandler := handlers.NewReceiptHandler(receiptService)
	stripeHandler := handlers.NewStripeHandler(stripeService)
	agentHandler := handlers.NewAgentHandler(agentService, scheduledTaskService, agentMessageRepo, agentUsageRepo)

	customerRepo := repository.NewCustomerRepository(db)
	customerService := service.NewCustomerService(customerRepo)
	customerHandler := handlers.NewCustomerHandler(customerService)

	invoiceSettingsRepo := repository.NewInvoiceSettingsRepository(db)
	invoiceSettingsService := service.NewInvoiceSettingsService(invoiceSettingsRepo)
	invoiceSettingsHandler := handlers.NewInvoiceSettingsHandler(invoiceSettingsService)

	invoiceRepo := repository.NewInvoiceRepository(db)
	invoiceLineRepo := repository.NewInvoiceLineRepository(db)
	invoiceService := service.NewInvoiceService(invoiceRepo, invoiceLineRepo, invoiceSettingsRepo, customerRepo, voucherService, lineItemRepo)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceService)
	invoicePDFHandler := handlers.NewInvoicePDFHandler(invoiceService, companyService)
	agentService.SetInvoiceEmailServices(invoiceService, companyService, emailService)
	invoiceEmailHandler := handlers.NewInvoiceEmailHandler(invoiceService, companyService, emailService)
	supportHandler := handlers.NewSupportHandler(emailService, companyService)
	vatPDFHandler := handlers.NewVATPDFHandler(reportService, companyService)
	sieImportService := service.NewSIEImportService(voucherService, accountService, lineItemService)
	sieImportHandler := handlers.NewSIEImportHandler(sieImportService)

	bankConnRepo := repository.NewBankConnectionRepository(db)
	bankAcctRepo := repository.NewBankAccountRepository(db)
	bankTxnRepo := repository.NewBankTransactionRepository(db)
	catRuleRepo := repository.NewCategorizationRuleRepository(db)
	tinkStateRepo := repository.NewTinkOAuthStateRepository(db)
	tinkClient := service.NewTinkClient(cfg.TinkClientID, cfg.TinkClientSecret, cfg.TinkCallbackURL)
	tinkService := service.NewTinkService(tinkClient, cfg.TinkEncryptionKey, bankConnRepo, bankAcctRepo, bankTxnRepo, catRuleRepo, tinkStateRepo, voucherService, lineItemService, cfg.FrontendURL)
	bankHandler := handlers.NewBankHandler(tinkService)

	authMiddleware := middleware.AuthMiddleware(jwtManager)
	companyMiddleware := middleware.CompanyMiddleware(companyRepo)

	router := gin.Default()
	router.ForwardedByClientIP = true
	if err := router.SetTrustedProxies([]string{"0.0.0.0/0"}); err != nil {
		log.Fatal("Failed to set trusted proxies:", err)
	}

	// Add security headers
	router.Use(middleware.SecurityHeadersMiddleware())

	// Add CORS middleware
	router.Use(middleware.CORSMiddleware())

	// Add rate limiting (100 req/min per IP)
	router.Use(middleware.RateLimitMiddleware())

	routes.SetupRoutes(router, userHandler, accountHandler, lineItemHandler, voucherHandler, authHandler, pdfHandler, reportHandler, sieHandler, agentHandler, companyHandler, receiptHandler, stripeHandler, customerHandler, invoiceSettingsHandler, invoiceHandler, invoicePDFHandler, invoiceEmailHandler, supportHandler, vatPDFHandler, sieImportHandler, bankHandler, authMiddleware, companyMiddleware)

	// Start the scheduler for recurring tasks
	schedulerService := service.NewSchedulerService(scheduledTaskService, agentService, invoiceService, func() []int {
		ids, _ := companyRepo.GetAllCompanyIDs()
		return ids
	})
	schedulerService.Start()
	defer schedulerService.Stop()

	log.Println("Starting server on", cfg.ServerPort)
	if err := router.Run(cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}