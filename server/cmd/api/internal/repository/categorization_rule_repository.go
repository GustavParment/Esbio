package repository

import (
	"esbio/api/internal/domain"
	"database/sql"
	"fmt"
)

type CategorizationRuleRepository interface {
	Upsert(rule *domain.CategorizationRule) error
	GetByCompanyID(companyID int) ([]*domain.CategorizationRule, error)
	FindMatch(companyID int, text string) (*domain.CategorizationRule, error)
	IncrementUseCount(ruleID int) error
	Delete(ruleID int) error
}

type categorizationRuleRepository struct {
	db *sql.DB
}

func NewCategorizationRuleRepository(db *sql.DB) CategorizationRuleRepository {
	return &categorizationRuleRepository{db: db}
}

func (r *categorizationRuleRepository) Upsert(rule *domain.CategorizationRule) error {
	query := `
		INSERT INTO bank_categorization_rules (company_id, match_pattern, match_type, account_no, description_template, tax_code, confidence)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT ON CONSTRAINT categorization_rules_company_pattern_unique
		DO UPDATE SET
			account_no = EXCLUDED.account_no,
			description_template = EXCLUDED.description_template,
			tax_code = EXCLUDED.tax_code,
			confidence = EXCLUDED.confidence,
			updated_at = CURRENT_TIMESTAMP
		RETURNING rule_id, created_at, updated_at
	`
	return r.db.QueryRow(query,
		rule.CompanyID, rule.MatchPattern, rule.MatchType,
		rule.AccountNo, rule.DescriptionTemplate, rule.TaxCode, rule.Confidence,
	).Scan(&rule.RuleID, &rule.CreatedAt, &rule.UpdatedAt)
}

func (r *categorizationRuleRepository) GetByCompanyID(companyID int) ([]*domain.CategorizationRule, error) {
	query := `
		SELECT rule_id, company_id, match_pattern, match_type, account_no, description_template, tax_code, confidence, use_count, created_at, updated_at
		FROM bank_categorization_rules
		WHERE company_id = $1
		ORDER BY use_count DESC, confidence DESC
	`
	rows, err := r.db.Query(query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categorization rules: %w", err)
	}
	defer rows.Close()

	var rules []*domain.CategorizationRule
	for rows.Next() {
		rule := &domain.CategorizationRule{}
		if err := rows.Scan(
			&rule.RuleID, &rule.CompanyID, &rule.MatchPattern, &rule.MatchType,
			&rule.AccountNo, &rule.DescriptionTemplate, &rule.TaxCode,
			&rule.Confidence, &rule.UseCount, &rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan categorization rule: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *categorizationRuleRepository) FindMatch(companyID int, text string) (*domain.CategorizationRule, error) {
	// Try exact match first, then contains match
	query := `
		SELECT rule_id, company_id, match_pattern, match_type, account_no, description_template, tax_code, confidence, use_count, created_at, updated_at
		FROM bank_categorization_rules
		WHERE company_id = $1
			AND (
				(match_type = 'exact' AND LOWER(match_pattern) = LOWER($2))
				OR (match_type = 'contains' AND LOWER($2) LIKE '%' || LOWER(match_pattern) || '%')
			)
		ORDER BY
			CASE match_type WHEN 'exact' THEN 0 ELSE 1 END,
			confidence DESC,
			use_count DESC
		LIMIT 1
	`
	rule := &domain.CategorizationRule{}
	err := r.db.QueryRow(query, companyID, text).Scan(
		&rule.RuleID, &rule.CompanyID, &rule.MatchPattern, &rule.MatchType,
		&rule.AccountNo, &rule.DescriptionTemplate, &rule.TaxCode,
		&rule.Confidence, &rule.UseCount, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // No match found, not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find matching rule: %w", err)
	}
	return rule, nil
}

func (r *categorizationRuleRepository) IncrementUseCount(ruleID int) error {
	query := `UPDATE bank_categorization_rules SET use_count = use_count + 1, updated_at = CURRENT_TIMESTAMP WHERE rule_id = $1`
	_, err := r.db.Exec(query, ruleID)
	return err
}

func (r *categorizationRuleRepository) Delete(ruleID int) error {
	query := `DELETE FROM bank_categorization_rules WHERE rule_id = $1`
	_, err := r.db.Exec(query, ruleID)
	return err
}
