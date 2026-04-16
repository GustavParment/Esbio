package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"esbio/api/internal/auth"
	"esbio/api/internal/domain"
	"esbio/api/internal/dto"
	"esbio/api/internal/repository"
	"esbio/api/internal/service"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func setCookie(c *gin.Context, name, value string, maxAge int) {
	setCookieWithOptions(c, name, value, maxAge, true)
}

func setReadableCookie(c *gin.Context, name, value string, maxAge int) {
	setCookieWithOptions(c, name, value, maxAge, false)
}

func setCookieWithOptions(c *gin.Context, name, value string, maxAge int, httpOnly bool) {
	sameSite := http.SameSiteNoneMode
	secure := true

	if os.Getenv("COOKIE_SAMESITE") == "lax" {
		sameSite = http.SameSiteLaxMode
	}
	if os.Getenv("COOKIE_SECURE") == "false" {
		secure = false
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     "/",
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: sameSite,
	})
}

type AuthHandler struct {
	userService  *service.UserService
	jwtManager   *auth.JWTManager
	emailService *service.EmailService
	userRepo     repository.UserRepository
	frontendURL  string
}

func NewAuthHandler(
	userService *service.UserService,
	jwtManager *auth.JWTManager,
	emailService *service.EmailService,
	userRepo repository.UserRepository,
	frontendURL string,
) *AuthHandler {
	return &AuthHandler{
		userService:  userService,
		jwtManager:   jwtManager,
		emailService: emailService,
		userRepo:     userRepo,
		frontendURL:  frontendURL,
	}
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Support both old (name) and new (first_name + last_name) format
	firstName := req.FirstName
	lastName := req.LastName
	if firstName == "" && lastName == "" && req.Name != "" {
		parts := splitName(req.Name)
		firstName = parts[0]
		lastName = parts[1]
	}

	user := &domain.User{
		FirstName:     firstName,
		LastName:      lastName,
		Email:         req.Email,
		PasswordHash:  req.Password,
		Role:          "Bookkeeper",
		EmailVerified: false,
	}

	if err := h.userService.CreateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate verification token
	token, err := generateToken()
	if err != nil {
		log.Printf("[ERROR] Failed to generate verification token: %v", err)
	} else {
		expires := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
		if err := h.userRepo.SetVerificationToken(user.UserID, token, expires); err != nil {
			log.Printf("[ERROR] Failed to save verification token: %v", err)
		} else {
			verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", h.frontendURL, token)
			if _, err := h.emailService.SendVerificationEmail(user.Email, firstName, verifyURL); err != nil {
				log.Printf("[ERROR] Failed to send verification email: %v", err)
			}
		}
	}

	jwtToken, err := h.jwtManager.GenerateToken(user.UserID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Set JWT as httpOnly cookie
	setCookie(c, "token", jwtToken, 604800)

	user.PasswordHash = ""

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully — check your email to verify",
		"user":    user,
	})
}

// VerifyEmail handles GET /auth/verify?token=xxx
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing token"})
		return
	}

	user, err := h.userRepo.VerifyEmail(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ogiltig eller utgången verifieringslänk"})
		return
	}

	// Send welcome email
	if _, err := h.emailService.SendWelcomeEmail(user.Email, user.FirstName); err != nil {
		log.Printf("[ERROR] Failed to send welcome email to %s: %v", user.Email, err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "e-post verifierad"})
}

// ForgotPassword handles POST /auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ange en giltig e-postadress"})
		return
	}

	// Always return success to avoid email enumeration
	token, err := generateToken()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "om kontot finns skickas ett mail"})
		return
	}

	expires := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	if err := h.userRepo.SetResetToken(req.Email, token, expires); err != nil {
		// User not found — still return success
		c.JSON(http.StatusOK, gin.H{"message": "om kontot finns skickas ett mail"})
		return
	}

	user, _ := h.userService.GetUserByEmail(req.Email)
	firstName := ""
	if user != nil {
		firstName = user.FirstName
	}

	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", h.frontendURL, token)
	if _, err := h.emailService.SendPasswordResetEmail(req.Email, firstName, resetURL); err != nil {
		log.Printf("[ERROR] Failed to send reset email to %s: %v", req.Email, err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "om kontot finns skickas ett mail"})
}

// ResetPassword handles POST /auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ogiltig begäran"})
		return
	}

	user, err := h.userRepo.GetUserByResetToken(req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ogiltig eller utgången återställningslänk"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		internalError(c, err)
		return
	}

	// Set the already-hashed password and clear the reset token.
	// We use the repo directly to avoid UserService.UpdateUser which would hash again.
	user.PasswordHash = string(hashedPassword)
	if err := h.userRepo.UpdatePassword(user.UserID, user.PasswordHash); err != nil {
		internalError(c, err)
		return
	}
	if err := h.userRepo.ClearResetToken(user.UserID); err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "lösenord uppdaterat"})
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if !user.EmailVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "Verifiera din e-postadress först — kolla din inbox."})
		return
	}

	token, err := h.jwtManager.GenerateToken(user.UserID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Set JWT as httpOnly cookie
	setCookie(c, "token", token, 604800)

	user.PasswordHash = ""

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"user":    user,
	})
}

// RefreshToken handles POST /auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Get token from cookie
	tokenString, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no token found"})
		return
	}

	newToken, err := h.jwtManager.RefreshToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	// Set new token as httpOnly cookie
	setCookie(c, "token", newToken, 604800)

	c.JSON(http.StatusOK, gin.H{
		"message": "token refreshed successfully",
	})
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear the cookie by setting maxAge to -1
	setCookie(c, "token", "", -1)

	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

// GetCurrentUser handles GET /auth/me (requires authentication)
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	user, err := h.userService.GetUserByID(userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	user.PasswordHash = ""

	c.JSON(http.StatusOK, user)
}

// DeleteAccount handles DELETE /auth/account — deletes the authenticated user and all their data
func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	// Require confirmation in request body
	var req struct {
		Confirm string `json:"confirm"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Confirm != "DELETE" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "send {\"confirm\": \"DELETE\"} to confirm account deletion"})
		return
	}

	if err := h.userService.DeleteUserAndData(userID.(int)); err != nil {
		internalError(c, err)
		return
	}

	// Clear auth cookies
	setCookie(c, "token", "", -1)
	setReadableCookie(c, "company_id", "", -1)

	c.JSON(http.StatusOK, gin.H{"message": "account deleted successfully"})
}

func splitName(name string) [2]string {
	parts := strings.SplitN(strings.TrimSpace(name), " ", 2)
	if len(parts) == 2 {
		return [2]string{parts[0], parts[1]}
	}
	return [2]string{parts[0], ""}
}
