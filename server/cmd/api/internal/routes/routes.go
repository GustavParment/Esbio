package routes

import (
	"cmd/api/internal/handlers"
	"cmd/api/internal/middleware"

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
	agentHandler *handlers.AgentHandler,
	authMiddleware gin.HandlerFunc) {

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.GET("/me", authMiddleware, authHandler.GetCurrentUser)
		}

		users := v1.Group("/users", authMiddleware)
		{
			users.POST("", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUserByID)
			users.GET("/email/:email", userHandler.GetUserByEmail)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		accounts := v1.Group("/accounts", authMiddleware)
		{
			accounts.POST("", accountHandler.CreateAccount)
			accounts.GET("", accountHandler.GetAllAccounts)
			accounts.GET("/:accountNo", accountHandler.GetAccountByNo)
			accounts.GET("/:accountNo/ledger", accountHandler.GetAccountLedger)
			accounts.GET("/group/:group", accountHandler.GetAccountsByGroup)
			accounts.PUT("/:accountNo", accountHandler.UpdateAccount)
			accounts.DELETE("/:accountNo", accountHandler.DeleteAccount)
		}

		lineItems := v1.Group("/lineitems", authMiddleware)
		{
			lineItems.POST("", lineItemHandler.CreateLineItem)
			lineItems.GET("/:id", lineItemHandler.GetLineItemByID)
			lineItems.GET("/voucher/:voucherId", lineItemHandler.GetLineItemsByVoucherID)
			lineItems.GET("/account/:accountNo", lineItemHandler.GetLineItemsByAccountNo)
			lineItems.PUT("/:id", lineItemHandler.UpdateLineItem)
			lineItems.DELETE("/:id", lineItemHandler.DeleteLineItem)
		}

		vouchers := v1.Group("/vouchers", authMiddleware)
		{
			vouchers.POST("", voucherHandler.CreateVoucher)
			vouchers.GET("", voucherHandler.GetAllVouchers)
			vouchers.GET("/periods", voucherHandler.GetAllPeriods)
			vouchers.GET("/:id", voucherHandler.GetVoucherByID)
			vouchers.GET("/period/:period", voucherHandler.GetVouchersByPeriod)
			vouchers.GET("/user/:userId", voucherHandler.GetVouchersByCreatedBy)
			vouchers.GET("/:id/validate", voucherHandler.ValidateVoucherBalance)
			vouchers.POST("/:id/correct", voucherHandler.CreateCorrectionVoucher)
			vouchers.POST("/:id/correct-with-changes", voucherHandler.CreateCorrectionWithChanges)
			vouchers.GET("/:id/pdf", pdfHandler.GenerateVoucherPDF)
			// Only Admin can update or delete vouchers
			vouchers.PUT("/:id", middleware.RequireRole("Admin"), voucherHandler.UpdateVoucher)
			vouchers.DELETE("/:id", middleware.RequireRole("Admin"), voucherHandler.DeleteVoucher)
		}

		reports := v1.Group("/reports", authMiddleware)
		{
			reports.GET("/income-statement", reportHandler.GetIncomeStatement)
			reports.GET("/balance-sheet", reportHandler.GetBalanceSheet)
			reports.GET("/vat", reportHandler.GetVATReport)
		}

		agent := v1.Group("/agent", authMiddleware)
		{
			agent.POST("/chat", agentHandler.Chat)
			agent.GET("/messages/:conversationId", agentHandler.GetMessages)
			agent.GET("/tasks", agentHandler.GetScheduledTasks)
			agent.PUT("/tasks/:id/toggle", agentHandler.ToggleScheduledTask)
			agent.DELETE("/tasks/:id", agentHandler.DeleteScheduledTask)
		}
	}
}
