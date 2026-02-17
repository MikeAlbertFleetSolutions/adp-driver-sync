package adp

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type DriverHomeAddress struct {
	EmployeeNumber string
	LastName       string
	FirstName      string
	Address1       string
	Address2       string
	City           string
	State          string
	ZIPCode        string
}

// OAuth2Token represents an OAuth2 access token
type OAuth2Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	ExpiresAt   time.Time
}

// ADPWorkerResponse represents the ADP Workforce Now API response
type ADPWorkerResponse struct {
	Workers []ADPWorker `json:"workers"`
}

// ADPWorkerID represents an ADP worker ID object
type ADPWorkerID struct {
	IDValue string `json:"idValue"`
}

// ADPWorker represents a worker from ADP Workforce Now
type ADPWorker struct {
	WorkerID         ADPWorkerID         `json:"workerId"`
	Person           ADPPerson           `json:"person"`
	WorkAssignments  []ADPWorkAssignment `json:"workAssignments"`
	CustomFieldGroup ADPCustomFieldGroup `json:"customFieldGroup"`
}

// ADPCustomFieldGroup contains custom fields defined in ADP
type ADPCustomFieldGroup struct {
	StringFields []ADPCustomStringField `json:"stringFields"`
	CodeFields   []ADPCustomCodeField   `json:"codeFields"`
}

// ADPCustomStringField represents a custom string field in ADP
type ADPCustomStringField struct {
	NameCode    ADPNameCode `json:"nameCode"`
	StringValue string      `json:"stringValue"`
}

// ADPCustomCodeField represents a custom code field in ADP
type ADPCustomCodeField struct {
	NameCode  ADPNameCode `json:"nameCode"`
	CodeValue string      `json:"codeValue"`
}

// ADPNameCode represents a nameCode object in ADP
type ADPNameCode struct {
	CodeValue string `json:"codeValue"`
	ShortName string `json:"shortName"`
}

// ADPPerson contains personal information
type ADPPerson struct {
	LegalName    ADPName         `json:"legalName"`
	LegalAddress ADPLegalAddress `json:"legalAddress"`
}

// ADPName contains name information
type ADPName struct {
	GivenName   string `json:"givenName"`
	FamilyName1 string `json:"familyName1"`
}

// ADPLegalAddress contains the home address from person.legalAddress
type ADPLegalAddress struct {
	LineOne                  string           `json:"lineOne"`
	LineTwo                  string           `json:"lineTwo,omitempty"`
	CityName                 string           `json:"cityName"`
	CountrySubdivisionLevel1 ADPCountrySubdiv `json:"countrySubdivisionLevel1"`
	PostalCode               string           `json:"postalCode"`
}

// ADPCountrySubdiv contains state/region info
type ADPCountrySubdiv struct {
	CodeValue string `json:"codeValue"`
	ShortName string `json:"shortName"`
}

// ADPAssignmentStatus contains assignment status details
type ADPAssignmentStatus struct {
	StatusCode ADPStatusCode `json:"statusCode"`
}

// ADPStatusCode represents the status code of a work assignment
type ADPStatusCode struct {
	CodeValue string `json:"codeValue"`
	ShortName string `json:"shortName"`
}

// ADPWorkAssignment contains work assignment details
type ADPWorkAssignment struct {
	ItemID            string              `json:"itemID"`
	PayrollFileNumber string              `json:"payrollFileNumber"`
	PrimaryIndicator  bool                `json:"primaryIndicator"`
	AssignmentStatus  ADPAssignmentStatus `json:"assignmentStatus"`
	CustomFieldGroup  ADPCustomFieldGroup `json:"customFieldGroup"`
}

// Client represents the ADP API client
type Client struct {
	clientID     string
	clientSecret string
	tokenURL     string
	baseURL      string
	httpClient   *http.Client
	oauth2Token  *OAuth2Token
}

// NewClient creates a new ADP API client with OAuth2 and client certificate
func NewClient(clientID, clientSecret, baseURL, certFile, keyFile string) (*Client, error) {
	// Load client certificate and private key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	// Configure TLS with client certificate
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenURL:     fmt.Sprintf("%s/auth/oauth/v2/token", baseURL),
		baseURL:      baseURL,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}, nil
}

// getAccessToken retrieves an OAuth2 access token
func (c *Client) getAccessToken(ctx context.Context) error {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", c.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var token OAuth2Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	c.oauth2Token = &token
	return nil
}

// ensureValidToken ensures we have a valid access token
func (c *Client) ensureValidToken(ctx context.Context) error {
	if c.oauth2Token == nil || time.Now().After(c.oauth2Token.ExpiresAt.Add(-5*time.Minute)) {
		return c.getAccessToken(ctx)
	}
	return nil
}

// GetWorkers retrieves all workers from ADP Workforce Now API with pagination
func (c *Client) GetWorkers(ctx context.Context) ([]ADPWorker, error) {
	var allWorkers []ADPWorker
	skip := 0
	pageSize := 100 // ADP max is typically 100

	for {
		if err := c.ensureValidToken(ctx); err != nil {
			return nil, fmt.Errorf("failed to get valid token: %w", err)
		}

		// ADP Workforce Now workers endpoint with pagination
		workersURL := fmt.Sprintf("%s/hr/v2/workers", c.baseURL)

		req, err := http.NewRequestWithContext(ctx, "GET", workersURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create workers request: %w", err)
		}

		// Add pagination parameters
		q := req.URL.Query()
		q.Add("$top", fmt.Sprintf("%d", pageSize))
		q.Add("$skip", fmt.Sprintf("%d", skip))
		req.URL.RawQuery = q.Encode()

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.oauth2Token.AccessToken))
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get workers: %w", err)
		}

		// Read the full response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("workers request failed with status %d: %s", resp.StatusCode, string(body))
		}

		var response ADPWorkerResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to decode workers response: %w", err)
		}

		if len(response.Workers) == 0 {
			break // no more workers
		}

		allWorkers = append(allWorkers, response.Workers...)
		log.Printf("Fetched %d workers from ADP (total so far: %d)", len(response.Workers), len(allWorkers))

		// If we got fewer than pageSize, we've reached the end
		if len(response.Workers) < pageSize {
			break
		}

		skip += pageSize
	}

	return allWorkers, nil
}

// isOverdriveSyncField checks if a field name matches "OVERDRIVE SYNC" by checking
// both codeValue and shortName, and also checks for partial/contains matches.
func isOverdriveSyncField(nameCode ADPNameCode) bool {
	for _, name := range []string{nameCode.CodeValue, nameCode.ShortName} {
		normalized := strings.ToUpper(strings.TrimSpace(name))
		if normalized == "OVERDRIVE SYNC" ||
			normalized == "OVERDRIVE_SYNC" ||
			normalized == "OVERDRIVESYNC" ||
			strings.Contains(normalized, "OVERDRIVE") {
			return true
		}
	}
	return false
}

// searchCustomFieldGroup looks for the OVERDRIVE SYNC value in a CustomFieldGroup.
func searchCustomFieldGroup(cfg ADPCustomFieldGroup) (string, bool) {
	for _, field := range cfg.StringFields {
		if isOverdriveSyncField(field.NameCode) {
			return strings.TrimSpace(field.StringValue), true
		}
	}
	for _, field := range cfg.CodeFields {
		if isOverdriveSyncField(field.NameCode) {
			return strings.TrimSpace(field.CodeValue), true
		}
	}
	return "", false
}

// getOverdriveSyncValue returns the value of the "OVERDRIVE SYNC" custom field for a worker.
// It searches at the worker level and at each work assignment level.
// Returns "" (blank) if the field is not present or has no value.
func getOverdriveSyncValue(worker ADPWorker) string {
	// 1) Check worker-level custom fields
	if val, found := searchCustomFieldGroup(worker.CustomFieldGroup); found {
		return val
	}
	// 2) Check work assignment-level custom fields
	for _, wa := range worker.WorkAssignments {
		if val, found := searchCustomFieldGroup(wa.CustomFieldGroup); found {
			return val
		}
	}
	return ""
}

// GetDriverHomeAddresses gets the driver home addresses from ADP Workforce Now
func (c *Client) GetDriverHomeAddresses() ([]DriverHomeAddress, error) {
	ctx := context.Background()

	workers, err := c.GetWorkers(ctx)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	var driverHomeAddresses []DriverHomeAddress
	skippedInactive := 0
	skippedOverdrive := 0

	for _, worker := range workers {
		// Find the primary work assignment to get the payrollFileNumber
		if len(worker.WorkAssignments) == 0 {
			continue
		}

		// Skip terminated/inactive workers - only sync workers with active assignments
		primaryAssignment := worker.WorkAssignments[0]
		statusCode := strings.ToUpper(primaryAssignment.AssignmentStatus.StatusCode.CodeValue)
		if statusCode != "A" { // "A" = Active
			skippedInactive++
			continue
		}

		// Use payrollFileNumber from the primary (first) work assignment as the employee number
		employeeNumber := primaryAssignment.PayrollFileNumber
		if employeeNumber == "" {
			continue
		}

		// Check the OVERDRIVE SYNC custom field
		// "No" = do NOT sync, Blank = OK to sync
		overdriveSyncValue := getOverdriveSyncValue(worker)
		if strings.EqualFold(overdriveSyncValue, "No") {
			skippedOverdrive++
			continue
		}

		// Get address from person.legalAddress
		address := worker.Person.LegalAddress

		driverHomeAddresses = append(driverHomeAddresses, DriverHomeAddress{
			EmployeeNumber: employeeNumber,
			LastName:       worker.Person.LegalName.FamilyName1,
			FirstName:      worker.Person.LegalName.GivenName,
			Address1:       address.LineOne,
			Address2:       address.LineTwo,
			City:           address.CityName,
			State:          address.CountrySubdivisionLevel1.CodeValue,
			ZIPCode:        address.PostalCode,
		})
	}

	log.Printf("ADP filter results: %d total workers, %d skipped (inactive/terminated), %d skipped (OVERDRIVE SYNC=No), %d eligible for sync",
		len(workers), skippedInactive, skippedOverdrive, len(driverHomeAddresses))

	return driverHomeAddresses, nil
}
