package service

import (
	"esbio/api/internal/repository"
	"fmt"
	"log"

	"github.com/stripe/stripe-go/v82"
	billingSession "github.com/stripe/stripe-go/v82/billingportal/session"
	checkoutSession "github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/subscription"
)

type StripeService struct {
	companyRepo   repository.CompanyRepository
	frontendURL   string
	webhookSecret string
}

func NewStripeService(secretKey string, webhookSecret string, frontendURL string, companyRepo repository.CompanyRepository) *StripeService {
	stripe.Key = secretKey
	return &StripeService{
		companyRepo:   companyRepo,
		frontendURL:   frontendURL,
		webhookSecret: webhookSecret,
	}
}

func (s *StripeService) GetWebhookSecret() string {
	return s.webhookSecret
}

// CreateCheckoutSession creates a Stripe Checkout session for a company
func (s *StripeService) CreateCheckoutSession(companyID int, priceID string, companyName string) (string, error) {
	// Get or create Stripe customer
	company, err := s.companyRepo.GetCompanyByID(companyID)
	if err != nil {
		return "", fmt.Errorf("failed to get company: %w", err)
	}

	var customerID string
	if company.StripeCustomerID != nil && *company.StripeCustomerID != "" {
		customerID = *company.StripeCustomerID
	} else {
		// Create new Stripe customer
		params := &stripe.CustomerParams{
			Name: stripe.String(companyName),
			Params: stripe.Params{
				Metadata: map[string]string{
					"company_id": fmt.Sprintf("%d", companyID),
				},
			},
		}
		c, err := customer.New(params)
		if err != nil {
			return "", fmt.Errorf("failed to create Stripe customer: %w", err)
		}
		customerID = c.ID

		// Save customer ID
		if err := s.companyRepo.UpdateStripeCustomerID(companyID, customerID); err != nil {
			log.Printf("Warning: failed to save stripe_customer_id for company %d: %v", companyID, err)
		}
	}

	// Create Checkout Session
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(s.frontendURL + "/upgrade/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(s.frontendURL + "/upgrade"),
		Params: stripe.Params{
			Metadata: map[string]string{
				"company_id": fmt.Sprintf("%d", companyID),
			},
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"company_id": fmt.Sprintf("%d", companyID),
			},
		},
	}

	session, err := checkoutSession.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	return session.URL, nil
}

// CreatePortalSession creates a Stripe Customer Portal session for managing subscription
func (s *StripeService) CreatePortalSession(companyID int) (string, error) {
	company, err := s.companyRepo.GetCompanyByID(companyID)
	if err != nil {
		return "", fmt.Errorf("failed to get company: %w", err)
	}

	if company.StripeCustomerID == nil || *company.StripeCustomerID == "" {
		return "", fmt.Errorf("company has no Stripe customer")
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  company.StripeCustomerID,
		ReturnURL: stripe.String(s.frontendURL + "/settings"),
	}

	session, err := billingSession.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create portal session: %w", err)
	}

	return session.URL, nil
}

// HandleSubscriptionEvent processes subscription changes from webhooks.
//
// Defense in depth: even though the webhook signature has already been
// verified by the handler, we re-fetch the subscription directly from the
// Stripe API to get the authoritative state. This means a forged event
// (e.g. if signature verification were somehow bypassed) cannot upgrade
// a plan, because Stripe will 404 on a non-existent subscription ID.
func (s *StripeService) HandleSubscriptionEvent(sub *stripe.Subscription) error {
	if sub == nil || sub.ID == "" {
		return fmt.Errorf("subscription event has no subscription ID")
	}

	// Re-fetch from Stripe — ignore whatever the webhook body contained
	authoritative, err := subscription.Get(sub.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to verify subscription %s with Stripe: %w", sub.ID, err)
	}

	companyIDStr, ok := authoritative.Metadata["company_id"]
	if !ok {
		return fmt.Errorf("subscription %s has no company_id in metadata", authoritative.ID)
	}

	var companyID int
	if _, err := fmt.Sscanf(companyIDStr, "%d", &companyID); err != nil {
		return fmt.Errorf("invalid company_id in subscription metadata: %s", companyIDStr)
	}

	// Map price to plan name. Reject unknown prices rather than silently
	// defaulting to a paid tier — a forged price_id must never upgrade.
	if authoritative.Items == nil || len(authoritative.Items.Data) == 0 || authoritative.Items.Data[0].Price == nil {
		return fmt.Errorf("subscription %s has no price items", authoritative.ID)
	}
	priceID := authoritative.Items.Data[0].Price.ID
	plan := s.priceToPlan(priceID)
	if plan == "" {
		return fmt.Errorf("subscription %s has unknown price_id %s — refusing to upgrade", authoritative.ID, priceID)
	}

	// Map Stripe status to our plan_status
	planStatus := "active"
	switch authoritative.Status {
	case stripe.SubscriptionStatusActive:
		planStatus = "active"
	case stripe.SubscriptionStatusPastDue:
		planStatus = "past_due"
	case stripe.SubscriptionStatusCanceled, stripe.SubscriptionStatusUnpaid:
		plan = "free"
		planStatus = "canceled"
	case stripe.SubscriptionStatusTrialing:
		planStatus = "trialing"
	}

	log.Printf("Stripe webhook: company=%d plan=%s status=%s subscription=%s", companyID, plan, planStatus, authoritative.ID)

	return s.companyRepo.UpdateSubscription(companyID, authoritative.ID, plan, planStatus)
}

// HandleCheckoutCompleted processes a completed checkout session
func (s *StripeService) HandleCheckoutCompleted(session *stripe.CheckoutSession) error {
	companyIDStr, ok := session.Metadata["company_id"]
	if !ok {
		return fmt.Errorf("no company_id in session metadata")
	}

	var companyID int
	if _, err := fmt.Sscanf(companyIDStr, "%d", &companyID); err != nil {
		return fmt.Errorf("invalid company_id: %s", companyIDStr)
	}

	if session.Subscription == nil {
		return fmt.Errorf("no subscription in checkout session")
	}

	// Fetch the full subscription from Stripe API to get price/items
	sub, err := subscription.Get(session.Subscription.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch subscription: %w", err)
	}

	return s.HandleSubscriptionEvent(sub)
}

// priceToPlan maps Stripe Price IDs to plan names. Returns an empty
// string for unknown prices — callers must reject, never default to a
// paid tier.
func (s *StripeService) priceToPlan(priceID string) string {
	switch priceID {
	// Test/sandbox prices
	case "price_1TKFEmF27XxV0OojLLuaOAfg":
		return "starter"
	case "price_1TKFFGF27XxV0Ooj3uaT41ho":
		return "growth"
	// Live/production prices
	case "price_1TKG9DFQgVJ3jY3cTQwpxKd7":
		return "starter"
	case "price_1TKG9HFQgVJ3jY3cQWf0Tq16":
		return "growth"
	default:
		log.Printf("ERROR: unknown Stripe price_id %s — refusing to map to a plan", priceID)
		return ""
	}
}
