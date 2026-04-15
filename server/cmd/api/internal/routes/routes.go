package routes

import (
	"esbio/api/internal/handlers"
	"esbio/api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	router *gin.Engine,
	userHandler *handlers.UserHandler,
	accountHandler *handlers.AccountHandler,
	lineItemHandler *handlers.LineItemHandler,
	voucherHandler *handlers.VoucherHandler,
	authHandler *handlers.AuthHandler,
	pdfHandler *handlers.PDFHandler,
	reportHandler *handlers.ReportHandler,
	sieHandler *handlers.SIEHandler,
	agentHandler *handlers.AgentHandler,
	companyHandler *handlers.CompanyHandler,
	receiptHandler *handlers.ReceiptHandler,
	stripeHandler *handlers.StripeHandler,
	customerHandler *handlers.CustomerHandler,
	invoiceSettingsHandler *handlers.InvoiceSettingsHandler,
	invoiceHandler *handlers.InvoiceHandler,
	invoicePDFHandler *handlers.InvoicePDFHandler,
	invoiceEmailHandler *handlers.InvoiceEmailHandler,
	supportHandler *handlers.SupportHandler,
	vatPDFHandler *handlers.VATPDFHandler,
	sieImportHandler *handlers.SIEImportHandler,
	authMiddleware gin.HandlerFunc,
	companyMiddleware gin.HandlerFunc) {

	v1 := router.Group("/api/v1")
	{
		authRateLimit := middleware.AuthRateLimitMiddleware()
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authRateLimit, authHandler.Register)
			auth.POST("/login", authRateLimit, authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.GET("/me", authMiddleware, authHandler.GetCurrentUser)
			auth.GET("/verify", authHandler.VerifyEmail)
			auth.POST("/forgot-password", authRateLimit, authHandler.ForgotPassword)
			auth.POST("/reset-password", authRateLimit, authHandler.ResetPassword)
			auth.DELETE("/account", authMiddleware, authHandler.DeleteAccount)
		}

		companies := v1.Group("/companies", authMiddleware)
		{
			companies.GET("", companyHandler.ListCompanies)
			companies.POST("", companyHandler.CreateCompany)
			companies.POST("/select", companyHandler.SelectCompany)
			companies.PUT("/:id", companyHandler.UpdateCompany)
			companies.DELETE("/:id", companyHandler.DeleteCompany)
		}

		users := v1.Group("/users", authMiddleware)
		{
			users.POST("", middleware.RequireRole("Admin"), userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUserByID)
			users.GET("/email/:email", userHandler.GetUserByEmail)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", middleware.RequireRole("Admin"), userHandler.DeleteUser)
		}

		accounts := v1.Group("/accounts", authMiddleware, companyMiddleware)
		{
			accounts.POST("", middleware.RequireRole("Admin", "Bookkeeper"), accountHandler.CreateAccount)
			accounts.GET("", accountHandler.GetAllAccounts)
			accounts.GET("/:accountNo", accountHandler.GetAccountByNo)
			accounts.GET("/:accountNo/ledger", accountHandler.GetAccountLedger)
			accounts.GET("/group/:group", accountHandler.GetAccountsByGroup)
			accounts.PUT("/:accountNo", middleware.RequireRole("Admin", "Bookkeeper"), accountHandler.UpdateAccount)
			accounts.DELETE("/:accountNo", middleware.RequireRole("Admin"), accountHandler.DeleteAccount)
		}

		lineItems := v1.Group("/lineitems", authMiddleware, companyMiddleware)
		{
			lineItems.POST("", middleware.RequireRole("Admin", "Bookkeeper"), lineItemHandler.CreateLineItem)
			lineItems.GET("/:id", lineItemHandler.GetLineItemByID)
			lineItems.GET("/voucher/:voucherId", lineItemHandler.GetLineItemsByVoucherID)
			lineItems.GET("/account/:accountNo", lineItemHandler.GetLineItemsByAccountNo)
			lineItems.PUT("/:id", middleware.RequireRole("Admin", "Bookkeeper"), lineItemHandler.UpdateLineItem)
			lineItems.DELETE("/:id", middleware.RequireRole("Admin"), lineItemHandler.DeleteLineItem)
		}

		vouchers := v1.Group("/vouchers", authMiddleware, companyMiddleware)
		{
			vouchers.POST("", voucherHandler.CreateVoucher)
			vouchers.GET("", voucherHandler.GetAllVouchers)
			vouchers.GET("/periods", voucherHandler.GetAllPeriods)
			vouchers.GET("/:id", voucherHandler.GetVoucherByID)
			vouchers.GET("/period/:period", voucherHandler.GetVouchersByPeriod)
			vouchers.GET("/company", voucherHandler.GetVouchersByCompanyID)
			vouchers.GET("/:id/validate", voucherHandler.ValidateVoucherBalance)
			vouchers.POST("/:id/correct", middleware.RequireRole("Admin"), voucherHandler.CreateCorrectionVoucher)
			vouchers.POST("/:id/correct-with-changes", middleware.RequireRole("Admin"), voucherHandler.CreateCorrectionWithChanges)
			vouchers.GET("/:id/pdf", pdfHandler.GenerateVoucherPDF)
			// Only Admin can update or delete vouchers
			vouchers.PUT("/:id", middleware.RequireRole("Admin"), voucherHandler.UpdateVoucher)
			vouchers.DELETE("/:id", middleware.RequireRole("Admin"), voucherHandler.DeleteVoucher)
		}

		reports := v1.Group("/reports", authMiddleware, companyMiddleware)
		{
			reports.GET("/income-statement", reportHandler.GetIncomeStatement)
			reports.GET("/balance-sheet", reportHandler.GetBalanceSheet)
			reports.GET("/vat", reportHandler.GetVATReport)
			reports.GET("/vat/pdf", vatPDFHandler.GenerateVATDeclarationPDF)
			reports.GET("/sie", sieHandler.ExportSIE)
			reports.POST("/sie/import/preview", sieImportHandler.PreviewSIE)
			reports.POST("/sie/import", sieImportHandler.ImportSIE)
		}

		receipts := v1.Group("/receipts", authMiddleware, companyMiddleware)
		{
			receipts.POST("/scan", middleware.RequireRole("Admin", "Bookkeeper"), receiptHandler.Scan)
		}

		agent := v1.Group("/agent", authMiddleware, companyMiddleware)
		{
			agent.POST("/chat", agentHandler.Chat)
			agent.POST("/stream", agentHandler.ChatStream)
			agent.GET("/messages/:conversationId", agentHandler.GetMessages)
			agent.GET("/tasks", agentHandler.GetScheduledTasks)
			agent.PUT("/tasks/:id/toggle", agentHandler.ToggleScheduledTask)
			agent.DELETE("/tasks/:id", agentHandler.DeleteScheduledTask)
			agent.GET("/usage", agentHandler.GetUsage)
		}

		invoices := v1.Group("/invoices", authMiddleware, companyMiddleware)
		{
			invoices.POST("", middleware.RequireRole("Admin", "Bookkeeper"), invoiceHandler.CreateInvoice)
			invoices.GET("", invoiceHandler.ListInvoices)
			invoices.GET("/settings", invoiceSettingsHandler.GetSettings)
			invoices.PUT("/settings", middleware.RequireRole("Admin", "Bookkeeper"), invoiceSettingsHandler.UpdateSettings)
			invoices.GET("/:id", invoiceHandler.GetInvoiceByID)
			invoices.GET("/:id/pdf", invoicePDFHandler.GenerateInvoicePDF)
			invoices.POST("/:id/finalize", middleware.RequireRole("Admin", "Bookkeeper"), invoiceHandler.FinalizeInvoice)
			invoices.POST("/:id/pay", middleware.RequireRole("Admin", "Bookkeeper"), invoiceHandler.MarkAsPaid)
			invoices.POST("/:id/cancel", middleware.RequireRole("Admin", "Bookkeeper"), invoiceHandler.CancelInvoice)
			invoices.POST("/:id/send-email", middleware.RequireRole("Admin", "Bookkeeper"), invoiceEmailHandler.SendInvoiceEmail)
			invoices.DELETE("/:id", middleware.RequireRole("Admin"), invoiceHandler.DeleteInvoice)
		}

		customers := v1.Group("/customers", authMiddleware, companyMiddleware)
		{
			customers.POST("", middleware.RequireRole("Admin", "Bookkeeper"), customerHandler.CreateCustomer)
			customers.GET("", customerHandler.ListCustomers)
			customers.GET("/search", customerHandler.SearchCustomers)
			customers.GET("/:id", customerHandler.GetCustomerByID)
			customers.PUT("/:id", middleware.RequireRole("Admin", "Bookkeeper"), customerHandler.UpdateCustomer)
			customers.DELETE("/:id", middleware.RequireRole("Admin"), customerHandler.DeleteCustomer)
		}

		v1.POST("/support", authMiddleware, companyMiddleware, supportHandler.SendSupportMessage)

		stripeGroup := v1.Group("/stripe")
		{
			// Checkout and portal require auth + company
			stripeGroup.POST("/checkout", authMiddleware, companyMiddleware, stripeHandler.CreateCheckoutSession)
			stripeGroup.POST("/portal", authMiddleware, companyMiddleware, stripeHandler.CreatePortalSession)
			// Webhook is called by Stripe — no auth
			stripeGroup.POST("/webhook", stripeHandler.HandleWebhook)
		}
	}
}
