package repository

import (
	"cmd/api/internal/domain"
	"database/sql"
	"fmt"
)

type CompanyRepository interface {
	CreateCompany(company *domain.Company) error
	GetCompanyByID(companyID int) (*domain.Company, error)
	GetCompaniesByUserID(userID int) ([]*domain.Company, error)
	UpdateCompany(company *domain.Company) error
	DeleteCompany(companyID int) error
	UserOwnsCompany(userID, companyID int) (bool, error)
}

type companyRepository struct {
	db *sql.DB
}

func NewCompanyRepository(db *sql.DB) CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) CreateCompany(company *domain.Company) error {
	query := `
		INSERT INTO companies (company_name, org_number, plan, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING company_id, created_at, updated_at
	`
	return r.db.QueryRow(query, company.CompanyName, company.OrgNumber, company.Plan, company.CreatedBy).
		Scan(&company.CompanyID, &company.CreatedAt, &company.UpdatedAt)
}

func (r *companyRepository) GetCompanyByID(companyID int) (*domain.Company, error) {
	query := `
		SELECT company_id, company_name, org_number, plan, created_by, created_at, updated_at
		FROM companies WHERE company_id = $1
	`
	c := &domain.Company{}
	var orgNumber sql.NullString
	err := r.db.QueryRow(query, companyID).Scan(
		&c.CompanyID, &c.CompanyName, &orgNumber, &c.Plan, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("company not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}
	c.OrgNumber = orgNumber.String
	return c, nil
}

func (r *companyRepository) GetCompaniesByUserID(userID int) ([]*domain.Company, error) {
	query := `
		SELECT company_id, company_name, org_number, plan, created_by, created_at, updated_at
		FROM companies WHERE created_by = $1 ORDER BY created_at
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get companies: %w", err)
	}
	defer rows.Close()

	var companies []*domain.Company
	for rows.Next() {
		c := &domain.Company{}
		var orgNumber sql.NullString
		if err := rows.Scan(&c.CompanyID, &c.CompanyName, &orgNumber, &c.Plan, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan company: %w", err)
		}
		c.OrgNumber = orgNumber.String
		companies = append(companies, c)
	}
	return companies, nil
}

func (r *companyRepository) UpdateCompany(company *domain.Company) error {
	query := `
		UPDATE companies SET company_name = $1, org_number = $2, plan = $3, updated_at = NOW()
		WHERE company_id = $4
	`
	_, err := r.db.Exec(query, company.CompanyName, company.OrgNumber, company.Plan, company.CompanyID)
	return err
}

func (r *companyRepository) DeleteCompany(companyID int) error {
	_, err := r.db.Exec("DELETE FROM companies WHERE company_id = $1", companyID)
	return err
}

func (r *companyRepository) UserOwnsCompany(userID, companyID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM companies WHERE company_id = $1 AND created_by = $2)", companyID, userID).Scan(&exists)
	return exists, err
}
