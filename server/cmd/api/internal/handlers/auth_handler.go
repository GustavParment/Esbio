package handlers

import (
	"esbio/api/internal/auth"
	"esbio/api/internal/domain"
	"esbio/api/internal/dto"
	"esbio/api/internal/service"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func setCookie(c *gin.Context, name, value string, maxAge int) {
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
		HttpOnly: true,
		SameSite: sameSite,
	})
}

type AuthHandler struct {
	userService *service.UserService
	jwtManager  *auth.JWTManager
}

func NewAuthHandler(
	userService *service.UserService, 
	jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtManager:  jwtManager,
	}
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
		FirstName:    firstName,
		LastName:     lastName,
		Email:        req.Email,
		PasswordHash: req.Password,
		Role:         "Bookkeeper",
	}

	if err := h.userService.CreateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"user":    user,
	})
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

func splitName(name string) [2]string {
	parts := strings.SplitN(strings.TrimSpace(name), " ", 2)
	if len(parts) == 2 {
		return [2]string{parts[0], parts[1]}
	}
	return [2]string{parts[0], ""}
}
