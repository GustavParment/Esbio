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
	normalized, err := normalizeOrgNumber(company.OrgNumber)
	if err != nil {
		return err
	}
	company.OrgNumber = normalized
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
	normalized, err := normalizeOrgNumber(company.OrgNumber)
	if err != nil {
		return err
	}
	company.OrgNumber = normalized
	return s.repo.UpdateCompany(company)
}

// normalizeOrgNumber returns the canonical "XXXXXX-XXXX" form, or "" if the
// input is empty (the field is optional). Returns an error for invalid input.
func normalizeOrgNumber(s string) (string, error) {
	digits := make([]byte, 0, 10)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			digits = append(digits, c)
		}
	}
	if len(digits) == 0 {
		return "", errors.New("organisationsnummer krävs")
	}
	if len(digits) != 10 {
		return "", errors.New("organisationsnummer måste vara 10 siffror")
	}
	if !luhnValid(digits) {
		return "", errors.New("ogiltigt organisationsnummer (kontrollsiffran stämmer inte)")
	}
	return string(digits[:6]) + "-" + string(digits[6:]), nil
}

func luhnValid(digits []byte) bool {
	sum := 0
	for i, b := range digits {
		n := int(b - '0')
		if i%2 == 0 {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
	}
	return sum%10 == 0
}

func (s *CompanyService) DeleteCompany(companyID int) error {
	return s.repo.DeleteCompany(companyID)
}

func (s *CompanyService) UserOwnsCompany(userID, companyID int) (bool, error) {
	return s.repo.UserOwnsCompany(userID, companyID)
}
