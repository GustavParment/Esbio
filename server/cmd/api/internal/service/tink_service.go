package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"esbio/api/internal/domain"
	"esbio/api/internal/repository"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

type TinkService struct {
	client          *TinkClient
	encryptionKey   []byte
	connRepo        repository.BankConnectionRepository
	accountRepo     repository.BankAccountRepository
	txnRepo         repository.BankTransactionRepository
	ruleRepo        repository.CategorizationRuleRepository
	stateRepo       repository.TinkOAuthStateRepository
	voucherService  *VoucherService
	lineItemService *LineItemService
	frontendURL     string
}

func NewTinkService(
	client *TinkClient,
	encryptionKeyHex string,
	connRepo repository.BankConnectionRepository,
	accountRepo repository.BankAccountRepository,
	txnRepo repository.BankTransactionRepository,
	ruleRepo repository.CategorizationRuleRepository,
	stateRepo repository.TinkOAuthStateRepository,
	voucherService *VoucherService,
	lineItemService *LineItemService,
	frontendURL string,
) *TinkService {
	var key []byte
	if encryptionKeyHex != "" {
		var err error
		key, err = hex.DecodeString(encryptionKeyHex)
		if err != nil || len(key) != 32 {
			log.Printf("WARNING: TINK_ENCRYPTION_KEY must be 64 hex chars (32 bytes). Token encryption disabled.")
			key = nil
		}
	}

	return &TinkService{
		client:          client,
		encryptionKey:   key,
		connRepo:        connRepo,
		accountRepo:     accountRepo,
		txnRepo:         txnRepo,
		ruleRepo:        ruleRepo,
		stateRepo:       stateRepo,
		voucherService:  voucherService,
		lineItemService: lineItemService,
		frontendURL:     frontendURL,
	}
}

// InitiateConnect generates a Tink Link URL for the user to connect their bank.
func (s *TinkService) InitiateConnect(companyID, userID int) (string, error) {
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	expiresAt := time.Now().Add(15 * time.Minute)
	if err := s.stateRepo.Create(state, companyID, userID, expiresAt); err != nil {
		return "", fmt.Errorf("failed to store oauth state: %w", err)
	}

	url := s.client.BuildConnectURL(state, "SE", "sv_SE")
	return url, nil
}

// HandleCallback processes the OAuth callback after the user connects their bank via Tink Link.
// Returns the companyID from the state token so the frontend can restore the company selection.
func (s *TinkService) HandleCallback(state, code string) (int, error) {
	companyID, userID, err := s.stateRepo.GetAndDelete(state)
	if err != nil {
		return 0, fmt.Errorf("invalid or expired state: %w", err)
	}

	tokenResp, err := s.client.ExchangeCode(code)
	if err != nil {
		return 0, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	// Encrypt the access token for storage
	encryptedToken, err := s.encryptToken(tokenResp.AccessToken)
	if err != nil {
		return 0, fmt.Errorf("failed to encrypt token: %w", err)
	}

	// Fetch accounts from Tink to get bank name and account details
	accounts, err := s.client.GetAccounts(tokenResp.AccessToken)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch accounts from Tink: %w", err)
	}

	bankName := "Bank"
	if len(accounts) > 0 {
		bankName = accounts[0].Name
	}

	// Default consent: 90 days from now (standard PSD2 consent)
	consentExpires := time.Now().Add(90 * 24 * time.Hour)

	conn := &domain.BankConnection{
		CompanyID:            companyID,
		CreatedBy:            userID,
		TinkUserID:           tokenResp.IDHint,
		BankName:             bankName,
		Status:               "active",
		AccessTokenEncrypted: &encryptedToken,
		ConsentExpiresAt:     consentExpires.Format(time.RFC3339),
	}

	if err := s.connRepo.Create(conn); err != nil {
		return 0, fmt.Errorf("failed to create bank connection: %w", err)
	}

	// Store each bank account
	for _, tinkAcct := range accounts {
		balance := parseTinkAmount(tinkAcct.Balances.Booked.Amount)
		now := time.Now().Format(time.RFC3339)

		bankAcct := &domain.BankAccount{
			ConnectionID:     conn.ConnectionID,
			CompanyID:        companyID,
			TinkAccountID:    tinkAcct.ID,
			AccountName:      tinkAcct.Name,
			HolderName:       nilIfEmpty(tinkAcct.HolderName),
			AccountType:      nilIfEmpty(tinkAcct.Type),
			IBAN:             tinkAccountIBAN(tinkAcct.Identifiers),
			AccountNumber:    tinkAccountNumber(tinkAcct.Identifiers),
			Currency:         tinkAcct.Balances.Booked.Amount.CurrencyCode,
			BalanceAmount:    &balance,
			BalanceUpdatedAt: &now,
		}
		if bankAcct.Currency == "" {
			bankAcct.Currency = "SEK"
		}

		if err := s.accountRepo.Create(bankAcct); err != nil {
			log.Printf("Failed to create bank account %s: %v", tinkAcct.ID, err)
		}
	}

	return companyID, nil
}

// GetConnections returns all bank connections for a company.
func (s *TinkService) GetConnections(companyID int) ([]*domain.BankConnection, error) {
	return s.connRepo.GetByCompanyID(companyID)
}

// GetAccounts returns all bank accounts for a company.
func (s *TinkService) GetAccounts(companyID int) ([]*domain.BankAccount, error) {
	return s.accountRepo.GetByCompanyID(companyID)
}

// GetTransactions returns bank transactions for a company, filtered by status.
func (s *TinkService) GetTransactions(companyID int, status string, limit, offset int) ([]*domain.BankTransaction, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return s.txnRepo.GetByCompanyID(companyID, status, limit, offset)
}

// SyncTransactions fetches new transactions from Tink and stores them.
func (s *TinkService) SyncTransactions(connectionID, companyID int) (int, error) {
	conn, err := s.connRepo.GetByID(connectionID)
	if err != nil {
		return 0, fmt.Errorf("connection not found: %w", err)
	}
	if conn.CompanyID != companyID {
		return 0, fmt.Errorf("connection does not belong to this company")
	}
	if conn.Status != "active" {
		return 0, fmt.Errorf("connection is not active (status: %s)", conn.Status)
	}
	if conn.AccessTokenEncrypted == nil {
		return 0, fmt.Errorf("no access token stored for this connection")
	}

	accessToken, err := s.decryptToken(*conn.AccessTokenEncrypted)
	if err != nil {
		return 0, fmt.Errorf("failed to decrypt access token: %w", err)
	}

	accounts, err := s.accountRepo.GetByConnectionID(connectionID)
	if err != nil {
		return 0, fmt.Errorf("failed to get bank accounts: %w", err)
	}

	totalNew := 0
	for _, acct := range accounts {
		tinkTxns, err := s.client.GetAllTransactions(accessToken, acct.TinkAccountID)
		if err != nil {
			log.Printf("Failed to fetch transactions for account %s: %v", acct.TinkAccountID, err)
			continue
		}

		var domainTxns []*domain.BankTransaction
		for _, tt := range tinkTxns {
			amount := parseTinkAmount(tt.Amount)
			txn := &domain.BankTransaction{
				BankAccountID:       acct.BankAccountID,
				CompanyID:           companyID,
				TinkTransactionID:   tt.ID,
				BookedDate:          tt.Dates.Booked,
				Amount:              amount,
				Currency:            tt.Amount.CurrencyCode,
				DescriptionOriginal: tt.Descriptions.Original,
				Status:              tt.Status,
			}
			if txn.Currency == "" {
				txn.Currency = "SEK"
			}
			if tt.Descriptions.Display != "" {
				txn.DescriptionDisplay = &tt.Descriptions.Display
			}
			if tt.MerchantData != nil {
				if tt.MerchantData.MerchantName != "" {
					txn.MerchantName = &tt.MerchantData.MerchantName
				}
				if tt.MerchantData.MerchantCategoryCode != "" {
					txn.MerchantCategoryCode = &tt.MerchantData.MerchantCategoryCode
				}
			}
			if tt.Categories != nil {
				txn.TinkCategoryID = &tt.Categories.PFM.ID
			}

			domainTxns = append(domainTxns, txn)
		}

		newCount, err := s.txnRepo.BulkUpsert(domainTxns)
		if err != nil {
			log.Printf("Failed to upsert transactions for account %s: %v", acct.TinkAccountID, err)
			continue
		}
		totalNew += newCount

		// Auto-suggest accounts for new pending transactions
		s.applyCategorization(companyID)

		// Update account balance
		if len(tinkTxns) > 0 {
			tinkAccounts, err := s.client.GetAccounts(accessToken)
			if err == nil {
				for _, ta := range tinkAccounts {
					if ta.ID == acct.TinkAccountID {
						balance := parseTinkAmount(ta.Balances.Booked.Amount)
						balF, _ := balance.Float64()
						_ = s.accountRepo.UpdateBalance(acct.BankAccountID, balF)
						break
					}
				}
			}
		}
	}

	_ = s.connRepo.UpdateLastSynced(connectionID)

	return totalNew, nil
}

// MapAccount maps a bank account to a BAS chart of accounts number.
func (s *TinkService) MapAccount(bankAccountID, companyID, accountNo int) error {
	acct, err := s.accountRepo.GetByID(bankAccountID)
	if err != nil {
		return fmt.Errorf("bank account not found: %w", err)
	}
	if acct.CompanyID != companyID {
		return fmt.Errorf("bank account does not belong to this company")
	}
	return s.accountRepo.UpdateMapping(bankAccountID, accountNo)
}

// ImportTransaction creates a voucher from a bank transaction.
func (s *TinkService) ImportTransaction(txnID, companyID, accountNo int, description string, taxCode int) (int, error) {
	txn, err := s.txnRepo.GetByID(txnID)
	if err != nil {
		return 0, fmt.Errorf("transaction not found: %w", err)
	}
	if txn.CompanyID != companyID {
		return 0, fmt.Errorf("transaction does not belong to this company")
	}
	if txn.ImportStatus == "booked" {
		return 0, fmt.Errorf("transaction is already imported")
	}

	// Get the mapped BAS account for the bank (e.g. 1930 Bankmedel)
	bankAcct, err := s.accountRepo.GetByID(txn.BankAccountID)
	if err != nil {
		return 0, fmt.Errorf("bank account not found: %w", err)
	}
	if bankAcct.MappedAccountNo == nil {
		return 0, fmt.Errorf("bank account is not mapped to a BAS account — map it first")
	}
	bankBASAccountNo := *bankAcct.MappedAccountNo

	// Derive period from booked_date (YYYY-MM)
	period := txn.BookedDate[:7]

	bookedTime, err := time.Parse("2006-01-02", txn.BookedDate)
	if err != nil {
		return 0, fmt.Errorf("invalid booked date: %w", err)
	}

	voucher := &domain.Voucher{
		Date:        domain.FlexibleDate{Time: bookedTime},
		Description: description,
		TotalAmount: txn.Amount.Abs(),
		Period:      period,
		CompanyID:   companyID,
	}

	if err := s.voucherService.CreateVoucher(voucher); err != nil {
		return 0, fmt.Errorf("failed to create voucher: %w", err)
	}

	// Create two line items: bank account ↔ target account
	var bankDebit, bankCredit, targetDebit, targetCredit decimal.Decimal
	if txn.Amount.IsPositive() {
		// Money received (income): debit bank, credit target
		bankDebit = txn.Amount
		targetCredit = txn.Amount
	} else {
		// Money spent (expense): credit bank, debit target
		absAmount := txn.Amount.Abs()
		bankCredit = absAmount
		targetDebit = absAmount
	}

	bankLine := &domain.LineItem{
		VoucherID:    voucher.VoucherID,
		AccountNo:    bankBASAccountNo,
		DebitAmount:  bankDebit,
		CreditAmount: bankCredit,
		TaxCode:      0, // Bank account lines have no VAT
	}
	if err := s.lineItemService.CreateLineItem(bankLine); err != nil {
		return 0, fmt.Errorf("failed to create bank line item: %w", err)
	}

	targetLine := &domain.LineItem{
		VoucherID:    voucher.VoucherID,
		AccountNo:    accountNo,
		DebitAmount:  targetDebit,
		CreditAmount: targetCredit,
		TaxCode:      taxCode,
	}
	if err := s.lineItemService.CreateLineItem(targetLine); err != nil {
		return 0, fmt.Errorf("failed to create target line item: %w", err)
	}

	// Mark transaction as booked
	voucherID := voucher.VoucherID
	if err := s.txnRepo.UpdateImportStatus(txnID, "booked", &voucherID); err != nil {
		return 0, fmt.Errorf("failed to update transaction status: %w", err)
	}

	// Learn: upsert categorization rule from this import
	matchText := txn.DescriptionOriginal
	if txn.MerchantName != nil && *txn.MerchantName != "" {
		matchText = *txn.MerchantName
	}
	if matchText != "" {
		rule := &domain.CategorizationRule{
			CompanyID:           companyID,
			MatchPattern:        matchText,
			MatchType:           "contains",
			AccountNo:           accountNo,
			DescriptionTemplate: &description,
			Confidence:          1.0,
		}
		if taxCode > 0 {
			rule.TaxCode = &taxCode
		}
		if err := s.ruleRepo.Upsert(rule); err != nil {
			log.Printf("Failed to upsert categorization rule: %v", err)
		}
	}

	return voucher.VoucherID, nil
}

// SkipTransaction marks a bank transaction as skipped.
func (s *TinkService) SkipTransaction(txnID, companyID int) error {
	txn, err := s.txnRepo.GetByID(txnID)
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}
	if txn.CompanyID != companyID {
		return fmt.Errorf("transaction does not belong to this company")
	}
	return s.txnRepo.UpdateImportStatus(txnID, "skipped", nil)
}

// DisconnectBank removes a bank connection and cascades to accounts/transactions.
func (s *TinkService) DisconnectBank(connectionID, companyID int) error {
	conn, err := s.connRepo.GetByID(connectionID)
	if err != nil {
		return fmt.Errorf("connection not found: %w", err)
	}
	if conn.CompanyID != companyID {
		return fmt.Errorf("connection does not belong to this company")
	}
	return s.connRepo.Delete(connectionID)
}

// applyCategorization runs categorization rules against pending transactions.
func (s *TinkService) applyCategorization(companyID int) {
	txns, err := s.txnRepo.GetPendingForReview(companyID)
	if err != nil {
		return
	}
	for _, txn := range txns {
		if txn.ImportStatus != "pending" {
			continue
		}
		// Try matching on merchant name first, then description
		matchText := txn.DescriptionOriginal
		if txn.MerchantName != nil && *txn.MerchantName != "" {
			matchText = *txn.MerchantName
		}
		rule, err := s.ruleRepo.FindMatch(companyID, matchText)
		if err != nil || rule == nil {
			continue
		}
		desc := matchText
		if rule.DescriptionTemplate != nil {
			desc = *rule.DescriptionTemplate
		}
		_ = s.txnRepo.UpdateSuggestion(txn.BankTransactionID, rule.AccountNo, desc)
		_ = s.ruleRepo.IncrementUseCount(rule.RuleID)
	}
}

// --- Encryption helpers ---

func (s *TinkService) encryptToken(plaintext string) (string, error) {
	if s.encryptionKey == nil {
		// No encryption key configured — store in plaintext (dev mode)
		return plaintext, nil
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *TinkService) decryptToken(encoded string) (string, error) {
	if s.encryptionKey == nil {
		// No encryption key — token stored in plaintext (dev mode)
		return encoded, nil
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode token: %w", err)
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt token: %w", err)
	}

	return string(plaintext), nil
}

// --- Amount parsing ---

// parseTinkAmount converts Tink's unscaledValue+scale to a decimal.Decimal.
// e.g. {unscaledValue: "12345", scale: "2"} → 123.45
func parseTinkAmount(a TinkCurrencyAmount) decimal.Decimal {
	if a.Value.UnscaledValue == "" {
		return decimal.Zero
	}

	unscaled, ok := new(big.Int).SetString(a.Value.UnscaledValue, 10)
	if !ok {
		return decimal.Zero
	}

	scale, _ := strconv.Atoi(a.Value.Scale)
	d := decimal.NewFromBigInt(unscaled, int32(-scale))
	return d
}

func tinkAccountIBAN(id TinkIdentifiers) *string {
	if id.IBAN != nil && id.IBAN.IBAN != "" {
		return &id.IBAN.IBAN
	}
	return nil
}

func tinkAccountNumber(id TinkIdentifiers) *string {
	if id.SortCode != nil && id.SortCode.AccountNumber != "" {
		return &id.SortCode.AccountNumber
	}
	if id.IBAN != nil && id.IBAN.BBAN != "" {
		return &id.IBAN.BBAN
	}
	return nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

