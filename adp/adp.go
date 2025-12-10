package adp

import (
	"context"
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

// ADPWorker represents a worker from ADP Workforce Now
type ADPWorker struct {
	WorkerID        string              `json:"workerId"`
	Person          ADPPerson           `json:"person"`
	WorkAssignments []ADPWorkAssignment `json:"workAssignments"`
}

// ADPPerson contains personal information
type ADPPerson struct {
	LegalName ADPName `json:"legalName"`
}

// ADPName contains name information
type ADPName struct {
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
}

// ADPWorkAssignment contains work assignment details including address
type ADPWorkAssignment struct {
	HomeWorkLocation ADPHomeLocation `json:"homeWorkLocation"`
}

// ADPHomeLocation contains home address information
type ADPHomeLocation struct {
	Address ADPAddress `json:"address"`
}

// ADPAddress contains the actual address fields
type ADPAddress struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	Region     string `json:"region"`
	PostalCode string `json:"postalCode"`
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

// NewClient creates a new ADP API client with OAuth2
func NewClient(clientID, clientSecret, baseURL string) (*Client, error) {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenURL:     fmt.Sprintf("%s/oauth/v2/token", baseURL),
		baseURL:      baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
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

// GetWorkers retrieves workers from ADP Workforce Now API
func (c *Client) GetWorkers(ctx context.Context) ([]ADPWorker, error) {
	if err := c.ensureValidToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to get valid token: %w", err)
	}

	// ADP Workforce Now workers endpoint
	workersURL := fmt.Sprintf("%s/hcm/v1/workers", c.baseURL)

	// Add query parameters to include address information
	params := url.Values{}
	params.Add("address", "true")                            // Include address data
	params.Add("$select", "workerId,person,workAssignments") // Select specific fields

	fullURL := fmt.Sprintf("%s?%s", workersURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create workers request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.oauth2Token.AccessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get workers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("workers request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response ADPWorkerResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode workers response: %w", err)
	}

	return response.Workers, nil
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

	for _, worker := range workers {
		// Extract employee number from worker ID (you may need to adjust this based on ADP's data)
		employeeNumber := extractEmployeeNumber(worker.WorkerID)

		// Get the primary work assignment (first one)
		if len(worker.WorkAssignments) == 0 {
			log.Printf("Worker %s has no work assignments", worker.WorkerID)
			continue
		}

		assignment := worker.WorkAssignments[0] // Use first assignment
		address := assignment.HomeWorkLocation.Address

		driverHomeAddresses = append(driverHomeAddresses, DriverHomeAddress{
			EmployeeNumber: employeeNumber,
			LastName:       worker.Person.LegalName.FamilyName,
			FirstName:      worker.Person.LegalName.GivenName,
			Address1:       address.Line1,
			Address2:       address.Line2,
			City:           address.City,
			State:          address.Region,
			ZIPCode:        address.PostalCode,
		})
	}

	return driverHomeAddresses, nil
}

// extractEmployeeNumber extracts employee number from ADP worker ID
// You may need to customize this based on how ADP stores employee numbers
func extractEmployeeNumber(workerID string) string {
	// ADP worker IDs might be in format "EMP001234" - extract the numeric part
	// Adjust this logic based on your ADP data structure
	return onlyNums(workerID)
}

func onlyNums(s string) string {
	bs := []byte(s)
	j := 0
	for _, b := range bs {
		if '0' <= b && b <= '9' {
			bs[j] = b
			j++
		}
	}
	return string(bs[:j])
}
