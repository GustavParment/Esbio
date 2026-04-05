package repository

import (
	"cmd/api/internal/domain"
	"database/sql"
	"fmt"
)

type AgentUsageRepository interface {
	SaveUsage(usage *domain.AgentUsage) error
	GetUsageSummary(userID int) (*domain.AgentUsageSummary, error)
}

type agentUsageRepository struct {
	db *sql.DB
}

func NewAgentUsageRepository(db *sql.DB) AgentUsageRepository {
	return &agentUsageRepository{db: db}
}

func (r *agentUsageRepository) SaveUsage(usage *domain.AgentUsage) error {
	query := `
		INSERT INTO agent_usage (user_id, prompt_tokens, completion_tokens, total_tokens)
		VALUES ($1, $2, $3, $4)
		RETURNING usage_id, created_at
	`
	return r.db.QueryRow(query, usage.UserID, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens).
		Scan(&usage.UsageID, &usage.CreatedAt)
}

func (r *agentUsageRepository) GetUsageSummary(userID int) (*domain.AgentUsageSummary, error) {
	query := `
		SELECT
			COALESCE(SUM(prompt_tokens), 0),
			COALESCE(SUM(completion_tokens), 0),
			COALESCE(SUM(total_tokens), 0),
			COUNT(*)
		FROM agent_usage
		WHERE user_id = $1
	`
	summary := &domain.AgentUsageSummary{}
	err := r.db.QueryRow(query, userID).Scan(
		&summary.TotalPromptTokens,
		&summary.TotalCompletionTokens,
		&summary.TotalTokens,
		&summary.RequestCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage summary: %w", err)
	}

	// Monthly usage
	monthQuery := `
		SELECT
			COALESCE(SUM(prompt_tokens), 0),
			COALESCE(SUM(completion_tokens), 0),
			COALESCE(SUM(total_tokens), 0),
			COUNT(*)
		FROM agent_usage
		WHERE user_id = $1
		  AND created_at >= date_trunc('month', CURRENT_DATE)
	`
	err = r.db.QueryRow(monthQuery, userID).Scan(
		&summary.MonthPromptTokens,
		&summary.MonthCompletionTokens,
		&summary.MonthTotalTokens,
		&summary.MonthRequestCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly usage: %w", err)
	}

	// Estimate cost (Gemini 2.5 Flash pricing in SEK, ~10.5 SEK/USD)
	usdCost := float64(summary.TotalPromptTokens)*0.15/1000000 + float64(summary.TotalCompletionTokens)*0.60/1000000
	summary.EstimatedCostSEK = usdCost * 10.5
	monthUsdCost := float64(summary.MonthPromptTokens)*0.15/1000000 + float64(summary.MonthCompletionTokens)*0.60/1000000
	summary.MonthEstimatedCostSEK = monthUsdCost * 10.5

	return summary, nil
}
