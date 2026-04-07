package handlers

import (
	"esbio/api/internal/domain"
	"esbio/api/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var user domain.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.CreateUser(&user); err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user created successfully",
		"user":    user,
	})
}

// GetUserByID handles GET /users/:id
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Only allow users to read their own profile, or Admin to read any
	requestingUserID, _ := c.Get("userID")
	role, _ := c.Get("role")
	if requestingUserID.(int) != userID && role.(string) != "Admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		notFoundError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUserByEmail handles GET /users/email/:email — Admin only
func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	role, _ := c.Get("role")
	requestingEmail, _ := c.Get("email")
	email := c.Param("email")

	// Only allow users to look up their own email, or Admin to look up any
	if requestingEmail.(string) != email && role.(string) != "Admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	user, err := h.userService.GetUserByEmail(email)
	if err != nil {
		notFoundError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser handles PUT /users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Only allow users to update their own profile, or Admin to update any
	requestingUserID, _ := c.Get("userID")
	role, _ := c.Get("role")
	isAdmin := role.(string) == "Admin"
	if requestingUserID.(int) != userID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Fetch existing user first
	existingUser, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Parse partial update
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Apply updates to existing user
	if v, ok := req["first_name"].(string); ok && v != "" {
		existingUser.FirstName = v
	}
	if v, ok := req["last_name"].(string); ok && v != "" {
		existingUser.LastName = v
	}
	if v, ok := req["email"].(string); ok && v != "" {
		existingUser.Email = v
	}
	if v, ok := req["company_name"].(string); ok {
		existingUser.CompanyName = v
	}
	if v, ok := req["org_number"].(string); ok {
		existingUser.OrgNumber = v
	}
	if v, ok := req["password"].(string); ok && v != "" {
		existingUser.PasswordHash = v
	}
	// Only Admin can change roles
	if v, ok := req["role"].(string); ok && v != "" {
		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "only admins can change user roles"})
			return
		}
		existingUser.Role = v
	}

	if err := h.userService.UpdateUser(existingUser); err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user updated successfully",
		"user":    existingUser,
	})
}

// DeleteUser handles DELETE /users/:id — Admin only
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Only Admin can delete users
	role, _ := c.Get("role")
	if role.(string) != "Admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can delete users"})
		return
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

func (h *UserHandler) Hello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
}