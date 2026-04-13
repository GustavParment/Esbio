package service

import (
	"esbio/api/internal/domain"
	"esbio/api/internal/repository"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type CustomerService struct {
	repo     repository.CustomerRepository
	validate *validator.Validate
}

func NewCustomerService(repo repository.CustomerRepository) *CustomerService {
	return &CustomerService{
		repo:     repo,
		validate: validator.New(),
	}
}

func (s *CustomerService) CreateCustomer(customer *domain.Customer) error {
	if err := s.validate.Struct(customer); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	if customer.Name == "" {
		return errors.New("customer name is required")
	}
	if customer.PaymentTermsDays <= 0 {
		customer.PaymentTermsDays = 30
	}
	if customer.Country == "" {
		customer.Country = "Sverige"
	}
	return s.repo.CreateCustomer(customer)
}

func (s *CustomerService) GetCustomerByID(customerID int) (*domain.Customer, error) {
	if customerID <= 0 {
		return nil, errors.New("invalid customer ID")
	}
	return s.repo.GetCustomerByID(customerID)
}

func (s *CustomerService) GetCustomersByCompanyID(companyID int) ([]*domain.Customer, error) {
	return s.repo.GetCustomersByCompanyID(companyID)
}

func (s *CustomerService) SearchCustomers(companyID int, query string) ([]*domain.Customer, error) {
	if query == "" {
		return s.repo.GetCustomersByCompanyID(companyID)
	}
	return s.repo.SearchCustomers(companyID, query)
}

func (s *CustomerService) UpdateCustomer(customer *domain.Customer) error {
	if customer.Name == "" {
		return errors.New("customer name is required")
	}
	return s.repo.UpdateCustomer(customer)
}

func (s *CustomerService) DeleteCustomer(customerID int) error {
	if customerID <= 0 {
		return errors.New("invalid customer ID")
	}
	return s.repo.DeleteCustomer(customerID)
}
