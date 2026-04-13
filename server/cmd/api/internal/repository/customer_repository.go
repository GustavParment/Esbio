package repository

import (
	"database/sql"
	"esbio/api/internal/domain"
	"fmt"
)

type CustomerRepository interface {
	CreateCustomer(customer *domain.Customer) error
	GetCustomerByID(customerID int) (*domain.Customer, error)
	GetCustomersByCompanyID(companyID int) ([]*domain.Customer, error)
	SearchCustomers(companyID int, query string) ([]*domain.Customer, error)
	UpdateCustomer(customer *domain.Customer) error
	DeleteCustomer(customerID int) error
}

type customerRepository struct {
	db *sql.DB
}

func NewCustomerRepository(db *sql.DB) CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) CreateCustomer(customer *domain.Customer) error {
	query := `
		INSERT INTO customers (company_id, name, org_number, vat_number, address_line1, address_line2,
			postal_code, city, country, email, phone, payment_terms_days, default_revenue_account, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING customer_id, created_at, updated_at
	`
	return r.db.QueryRow(query,
		customer.CompanyID, customer.Name, customer.OrgNumber, customer.VATNumber,
		customer.AddressLine1, customer.AddressLine2, customer.PostalCode, customer.City,
		customer.Country, customer.Email, customer.Phone, customer.PaymentTermsDays,
		customer.DefaultRevenueAccount, customer.Notes,
	).Scan(&customer.CustomerID, &customer.CreatedAt, &customer.UpdatedAt)
}

func (r *customerRepository) GetCustomerByID(customerID int) (*domain.Customer, error) {
	query := `
		SELECT customer_id, company_id, name, org_number, vat_number, address_line1, address_line2,
			postal_code, city, country, email, phone, payment_terms_days, default_revenue_account,
			notes, created_at, updated_at
		FROM customers WHERE customer_id = $1
	`
	c := &domain.Customer{}
	var orgNumber, vatNumber, addr1, addr2, postalCode, city, country, email, phone, notes sql.NullString
	var defaultAccount sql.NullInt64
	err := r.db.QueryRow(query, customerID).Scan(
		&c.CustomerID, &c.CompanyID, &c.Name, &orgNumber, &vatNumber,
		&addr1, &addr2, &postalCode, &city, &country, &email, &phone,
		&c.PaymentTermsDays, &defaultAccount, &notes, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	if orgNumber.Valid {
		c.OrgNumber = orgNumber.String
	}
	if vatNumber.Valid {
		c.VATNumber = vatNumber.String
	}
	if addr1.Valid {
		c.AddressLine1 = addr1.String
	}
	if addr2.Valid {
		c.AddressLine2 = addr2.String
	}
	if postalCode.Valid {
		c.PostalCode = postalCode.String
	}
	if city.Valid {
		c.City = city.String
	}
	if country.Valid {
		c.Country = country.String
	}
	if email.Valid {
		c.Email = email.String
	}
	if phone.Valid {
		c.Phone = phone.String
	}
	if notes.Valid {
		c.Notes = notes.String
	}
	if defaultAccount.Valid {
		val := int(defaultAccount.Int64)
		c.DefaultRevenueAccount = &val
	}
	return c, nil
}

func (r *customerRepository) GetCustomersByCompanyID(companyID int) ([]*domain.Customer, error) {
	query := `
		SELECT customer_id, company_id, name, COALESCE(org_number, ''), COALESCE(email, ''),
			COALESCE(phone, ''), COALESCE(city, ''), payment_terms_days, created_at, updated_at
		FROM customers WHERE company_id = $1 ORDER BY name
	`
	rows, err := r.db.Query(query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}
	defer rows.Close()

	var customers []*domain.Customer
	for rows.Next() {
		c := &domain.Customer{}
		if err := rows.Scan(
			&c.CustomerID, &c.CompanyID, &c.Name, &c.OrgNumber, &c.Email,
			&c.Phone, &c.City, &c.PaymentTermsDays, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, c)
	}
	return customers, nil
}

func (r *customerRepository) SearchCustomers(companyID int, query string) ([]*domain.Customer, error) {
	sqlQuery := `
		SELECT customer_id, company_id, name, COALESCE(org_number, ''), COALESCE(email, ''),
			COALESCE(phone, ''), COALESCE(city, ''), payment_terms_days, created_at, updated_at
		FROM customers
		WHERE company_id = $1 AND (name ILIKE $2 OR org_number ILIKE $2 OR email ILIKE $2)
		ORDER BY name
	`
	pattern := "%" + query + "%"
	rows, err := r.db.Query(sqlQuery, companyID, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}
	defer rows.Close()

	var customers []*domain.Customer
	for rows.Next() {
		c := &domain.Customer{}
		if err := rows.Scan(
			&c.CustomerID, &c.CompanyID, &c.Name, &c.OrgNumber, &c.Email,
			&c.Phone, &c.City, &c.PaymentTermsDays, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, c)
	}
	return customers, nil
}

func (r *customerRepository) UpdateCustomer(customer *domain.Customer) error {
	query := `
		UPDATE customers SET name = $1, org_number = $2, vat_number = $3, address_line1 = $4,
			address_line2 = $5, postal_code = $6, city = $7, country = $8, email = $9, phone = $10,
			payment_terms_days = $11, default_revenue_account = $12, notes = $13, updated_at = NOW()
		WHERE customer_id = $14
	`
	_, err := r.db.Exec(query,
		customer.Name, customer.OrgNumber, customer.VATNumber, customer.AddressLine1,
		customer.AddressLine2, customer.PostalCode, customer.City, customer.Country,
		customer.Email, customer.Phone, customer.PaymentTermsDays, customer.DefaultRevenueAccount,
		customer.Notes, customer.CustomerID,
	)
	return err
}

func (r *customerRepository) DeleteCustomer(customerID int) error {
	_, err := r.db.Exec("DELETE FROM customers WHERE customer_id = $1", customerID)
	return err
}
