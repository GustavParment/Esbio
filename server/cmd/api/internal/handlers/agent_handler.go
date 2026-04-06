package handlers

import (
	"esbio/api/internal/domain"
	"esbio/api/internal/repository"
	"esbio/api/internal/service"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	agentService   *service.AgentService
	taskService    *service.ScheduledTaskService
	messageRepo    repository.AgentMessageRepository
	usageRepo      repository.AgentUsageRepository
}

func NewAgentHandler(
	agentService *service.AgentService,
	taskService *service.ScheduledTaskService,
	messageRepo repository.AgentMessageRepository,
	usageRepo repository.AgentUsageRepository,
) *AgentHandler {
	return &AgentHandler{
		agentService: agentService,
		taskService:  taskService,
		messageRepo:  messageRepo,
		usageRepo:    usageRepo,
	}
}

// Chat handles POST /agent/chat
func (h *AgentHandler) Chat(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected"})
		return
	}

	var req struct {
		Message        string `json:"message" binding:"required"`
		ConversationID string `json:"conversation_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	uid := userID.(int)
	cid := companyID.(int)

	// Generate conversation ID if not provided
	if req.ConversationID == "" {
		req.ConversationID = fmt.Sprintf("conv-%d-%d", uid, time.Now().UnixNano())
	}

	// Load conversation history
	history, _ := h.messageRepo.GetMessagesByConversation(req.ConversationID)
	var historyMessages []domain.AgentMessage
	for _, msg := range history {
		historyMessages = append(historyMessages, *msg)
	}

	// Save user message
	userMsg := &domain.AgentMessage{
		UserID:         uid,
		CompanyID:      cid,
		ConversationID: req.ConversationID,
		Role:           "user",
		Content:        req.Message,
	}
	h.messageRepo.SaveMessage(userMsg)

	// Get agent response with conversation history
	response, err := h.agentService.Chat(uid, cid, req.ConversationID, req.Message, historyMessages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Save assistant message
	assistantMsg := &domain.AgentMessage{
		UserID:         uid,
		CompanyID:      cid,
		ConversationID: req.ConversationID,
		Role:           "assistant",
		Content:        response,
	}
	h.messageRepo.SaveMessage(assistantMsg)

	c.JSON(http.StatusOK, gin.H{
		"response":        response,
		"conversation_id": req.ConversationID,
	})
}

// GetMessages handles GET /agent/messages/:conversationId
func (h *AgentHandler) GetMessages(c *gin.Context) {
	conversationID := c.Param("conversationId")
	messages, err := h.messageRepo.GetMessagesByConversation(conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if messages == nil {
		messages = []*domain.AgentMessage{}
	}
	c.JSON(http.StatusOK, messages)
}

// GetScheduledTasks handles GET /agent/tasks
func (h *AgentHandler) GetScheduledTasks(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected"})
		return
	}

	tasks, err := h.taskService.GetTasksByCompanyID(companyID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tasks == nil {
		tasks = []*domain.ScheduledTask{}
	}
	c.JSON(http.StatusOK, tasks)
}

// ToggleScheduledTask handles PUT /agent/tasks/:id/toggle
func (h *AgentHandler) ToggleScheduledTask(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	if err := h.taskService.ToggleTask(taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	task, _ := h.taskService.GetTaskByID(taskID)
	c.JSON(http.StatusOK, task)
}

// DeleteScheduledTask handles DELETE /agent/tasks/:id
func (h *AgentHandler) DeleteScheduledTask(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	if err := h.taskService.DeleteTask(taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task deleted"})
}

// GetUsage handles GET /agent/usage
func (h *AgentHandler) GetUsage(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected"})
		return
	}

	summary, err := h.usageRepo.GetUsageSummary(companyID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}
