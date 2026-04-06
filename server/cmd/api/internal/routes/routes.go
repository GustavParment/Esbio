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
			users.POST("", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUserByID)
			users.GET("/email/:email", userHandler.GetUserByEmail)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		accounts := v1.Group("/accounts", authMiddleware, companyMiddleware)
		{
			accounts.POST("", accountHandler.CreateAccount)
			accounts.GET("", accountHandler.GetAllAccounts)
			accounts.GET("/:accountNo", accountHandler.GetAccountByNo)
			accounts.GET("/:accountNo/ledger", accountHandler.GetAccountLedger)
			accounts.GET("/group/:group", accountHandler.GetAccountsByGroup)
			accounts.PUT("/:accountNo", accountHandler.UpdateAccount)
			accounts.DELETE("/:accountNo", accountHandler.DeleteAccount)
		}

		lineItems := v1.Group("/lineitems", authMiddleware, companyMiddleware)
		{
			lineItems.POST("", lineItemHandler.CreateLineItem)
			lineItems.GET("/:id", lineItemHandler.GetLineItemByID)
			lineItems.GET("/voucher/:voucherId", lineItemHandler.GetLineItemsByVoucherID)
			lineItems.GET("/account/:accountNo", lineItemHandler.GetLineItemsByAccountNo)
			lineItems.PUT("/:id", lineItemHandler.UpdateLineItem)
			lineItems.DELETE("/:id", lineItemHandler.DeleteLineItem)
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
			vouchers.POST("/:id/correct", voucherHandler.CreateCorrectionVoucher)
			vouchers.POST("/:id/correct-with-changes", voucherHandler.CreateCorrectionWithChanges)
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
			reports.GET("/sie", sieHandler.ExportSIE)
		}

		agent := v1.Group("/agent", authMiddleware, companyMiddleware)
		{
			agent.POST("/chat", agentHandler.Chat)
			agent.GET("/messages/:conversationId", agentHandler.GetMessages)
			agent.GET("/tasks", agentHandler.GetScheduledTasks)
			agent.PUT("/tasks/:id/toggle", agentHandler.ToggleScheduledTask)
			agent.DELETE("/tasks/:id", agentHandler.DeleteScheduledTask)
			agent.GET("/usage", agentHandler.GetUsage)
		}
	}
}
