package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const tinkBaseURL = "https://api.tink.com"

// TinkClient wraps HTTP calls to the Tink API
type TinkClient struct {
	clientID     string
	clientSecret string
	callbackURL  string
	httpClient   *http.Client
}

// NewTinkClient creates a new Tink API client
func NewTinkClient(clientID, clientSecret, callbackURL string) *TinkClient {
	return &TinkClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		callbackURL:  callbackURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// --- Tink API response types ---

type TinkTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	IDHint       string `json:"id_hint"`
}

type TinkAccount struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Type          string             `json:"type"`
	IBAN          string             `json:"iban"`
	AccountNumber TinkAccountNumber  `json:"identifiers"`
	Balances      TinkBalances       `json:"balances"`
}

type TinkAccountNumber struct {
	SortCode      string `json:"sortCode"`
	AccountNumber string `json:"accountNumber"`
	IBAN          string `json:"iban"`
}

type TinkBalances struct {
	Booked TinkAmount `json:"booked"`
}

type TinkAmount struct {
	Amount   TinkCurrencyAmount `json:"amount"`
}

type TinkCurrencyAmount struct {
	Value        string `json:"unscaledValue"`
	Scale        int    `json:"scale"`
	CurrencyCode string `json:"currencyCode"`
}

type TinkAccountsResponse struct {
	Accounts      []TinkAccount `json:"accounts"`
	NextPageToken string        `json:"nextPageToken"`
}

type TinkTransaction struct {
	ID           string              `json:"id"`
	AccountID    string              `json:"accountId"`
	Amount       TinkCurrencyAmount  `json:"amount"`
	Descriptions TinkDescriptions    `json:"descriptions"`
	Dates        TinkDates           `json:"dates"`
	Status       string              `json:"status"`
	MerchantData *TinkMerchantData   `json:"merchantInformation,omitempty"`
	Categories   *TinkCategories     `json:"categories,omitempty"`
}

type TinkDescriptions struct {
	Original string `json:"original"`
	Display  string `json:"display"`
}

type TinkDates struct {
	Booked string `json:"booked"`
	Value  string `json:"value"`
}

type TinkMerchantData struct {
	MerchantName         string `json:"merchantName"`
	MerchantCategoryCode string `json:"merchantCategoryCode"`
}

type TinkCategories struct {
	PFM TinkPFMCategory `json:"pfm"`
}

type TinkPFMCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TinkTransactionsResponse struct {
	Transactions  []TinkTransaction `json:"transactions"`
	NextPageToken string            `json:"nextPageToken"`
}

// --- Public methods ---

// BuildConnectURL builds the Tink Link URL for connecting a bank account
func (c *TinkClient) BuildConnectURL(state, market, locale string) string {
	params := url.Values{
		"client_id":    {c.clientID},
		"redirect_uri": {c.callbackURL},
		"market":       {market},
		"locale":       {locale},
		"state":        {state},
	}
	return fmt.Sprintf("https://link.tink.com/1.0/transactions/connect-accounts?%s", params.Encode())
}

// ExchangeCode exchanges an authorization code for an access token
func (c *TinkClient) ExchangeCode(code string) (*TinkTokenResponse, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
		"grant_type":    {"authorization_code"},
	}

	resp, err := c.httpClient.Post(
		tinkBaseURL+"/api/v1/oauth/token",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tink token exchange failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp TinkTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// GetAccounts fetches all bank accounts for a Tink user
func (c *TinkClient) GetAccounts(accessToken string) ([]TinkAccount, error) {
	var allAccounts []TinkAccount
	pageToken := ""

	for {
		u := tinkBaseURL + "/data/v2/accounts"
		if pageToken != "" {
			u += "?pageToken=" + url.QueryEscape(pageToken)
		}

		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get accounts: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("tink get accounts failed (status %d): %s", resp.StatusCode, string(body))
		}

		var accountsResp TinkAccountsResponse
		if err := json.NewDecoder(resp.Body).Decode(&accountsResp); err != nil {
			return nil, fmt.Errorf("failed to decode accounts: %w", err)
		}

		allAccounts = append(allAccounts, accountsResp.Accounts...)

		if accountsResp.NextPageToken == "" {
			break
		}
		pageToken = accountsResp.NextPageToken
	}

	return allAccounts, nil
}

// GetTransactions fetches transactions for a specific account, paginated
func (c *TinkClient) GetTransactions(accessToken, accountID string, pageToken string) (*TinkTransactionsResponse, error) {
	params := url.Values{
		"accountIdIn": {accountID},
		"pageSize":    {"100"},
	}
	if pageToken != "" {
		params.Set("pageToken", pageToken)
	}

	u := tinkBaseURL + "/data/v2/transactions?" + params.Encode()

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tink get transactions failed (status %d): %s", resp.StatusCode, string(body))
	}

	var txnResp TinkTransactionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&txnResp); err != nil {
		return nil, fmt.Errorf("failed to decode transactions: %w", err)
	}

	return &txnResp, nil
}

// GetAllTransactions fetches all transactions for an account, handling pagination
func (c *TinkClient) GetAllTransactions(accessToken, accountID string) ([]TinkTransaction, error) {
	var all []TinkTransaction
	pageToken := ""

	for {
		resp, err := c.GetTransactions(accessToken, accountID, pageToken)
		if err != nil {
			return nil, err
		}

		all = append(all, resp.Transactions...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return all, nil
}
