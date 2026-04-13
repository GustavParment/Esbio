package service

import (
	"esbio/api/internal/domain"
	"esbio/api/internal/repository"
)

type InvoiceSettingsService struct {
	repo repository.InvoiceSettingsRepository
}

func NewInvoiceSettingsService(repo repository.InvoiceSettingsRepository) *InvoiceSettingsService {
	return &InvoiceSettingsService{repo: repo}
}

func (s *InvoiceSettingsService) GetByCompanyID(companyID int) (*domain.InvoiceSettings, error) {
	return s.repo.GetByCompanyID(companyID)
}

func (s *InvoiceSettingsService) Update(settings *domain.InvoiceSettings) error {
	return s.repo.Upsert(settings)
}
