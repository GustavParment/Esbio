package handlers

import (
	"encoding/json"
	"esbio/api/internal/service"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

type StripeHandler struct {
	stripeService *service.StripeService
}

func NewStripeHandler(stripeService *service.StripeService) *StripeHandler {
	return &StripeHandler{stripeService: stripeService}
}

// CreateCheckoutSession handles POST /stripe/checkout
func (h *StripeHandler) CreateCheckoutSession(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected"})
		return
	}

	var req struct {
		PriceID string `json:"price_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "price_id is required"})
		return
	}

	// Get company name for Stripe customer
	companyName := "Esbio Company"
	if name, exists := c.Get("companyName"); exists {
		companyName = name.(string)
	}

	url, err := h.stripeService.CreateCheckoutSession(companyID.(int), req.PriceID, companyName)
	if err != nil {
		log.Printf("Stripe checkout error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create checkout session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// CreatePortalSession handles POST /stripe/portal
func (h *StripeHandler) CreatePortalSession(c *gin.Context) {
	companyID, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no company selected"})
		return
	}

	url, err := h.stripeService.CreatePortalSession(companyID.(int))
	if err != nil {
		log.Printf("Stripe portal error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create portal session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// HandleWebhook handles POST /stripe/webhook (no auth — called by Stripe)
func (h *StripeHandler) HandleWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	webhookSecret := h.stripeService.GetWebhookSecret()

	var event stripe.Event

	if webhookSecret != "" {
		event, err = webhook.ConstructEvent(body, c.GetHeader("Stripe-Signature"), webhookSecret)
		if err != nil {
			log.Printf("Stripe webhook signature verification failed: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
			return
		}
	} else {
		// No webhook secret configured — parse without verification (dev only)
		if err := json.Unmarshal(body, &event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event"})
			return
		}
	}

	switch event.Type {
	case "customer.subscription.created",
		"customer.subscription.updated",
		"customer.subscription.deleted":

		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			log.Printf("Failed to parse subscription from webhook: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription data"})
			return
		}
		if err := h.stripeService.HandleSubscriptionEvent(&sub); err != nil {
			log.Printf("Failed to handle subscription event: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process event"})
			return
		}

	default:
		log.Printf("Unhandled Stripe event type: %s", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}
