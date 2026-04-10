package service

import (
	"esbio/api/internal/domain"
	"esbio/api/internal/repository"
	"errors"
)

type CompanyService struct {
	repo repository.CompanyRepository
}

func NewCompanyService(repo repository.CompanyRepository) *CompanyService {
	return &CompanyService{repo: repo}
}

func (s *CompanyService) CreateCompany(company *domain.Company) error {
	if company.CompanyName == "" {
		return errors.New("company name is required")
	}
	if company.CreatedBy <= 0 {
		return errors.New("created_by is required")
	}
	if company.Plan == "" {
		company.Plan = "free"
	}
	return s.repo.CreateCompany(company)
}

func (s *CompanyService) GetCompanyByID(companyID int) (*domain.Company, error) {
	return s.repo.GetCompanyByID(companyID)
}

func (s *CompanyService) GetCompanyByOrgNumber(orgNumber string) (*domain.Company, error) {
	return s.repo.GetCompanyByOrgNumber(orgNumber)
}

func (s *CompanyService) GetCompaniesByUserID(userID int) ([]*domain.Company, error) {
	return s.repo.GetCompaniesByUserID(userID)
}

func (s *CompanyService) UpdateCompany(company *domain.Company) error {
	if company.CompanyName == "" {
		return errors.New("company name is required")
	}
	return s.repo.UpdateCompany(company)
}

func (s *CompanyService) DeleteCompany(companyID int) error {
	return s.repo.DeleteCompany(companyID)
}

func (s *CompanyService) UserOwnsCompany(userID, companyID int) (bool, error) {
	return s.repo.UserOwnsCompany(userID, companyID)
}
