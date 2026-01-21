package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// Block represents a Letta memory block
type Block struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

// Client is a Letta API client
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new Letta API client
func NewClient(baseURL string) (*Client, error) {
	apiKey := os.Getenv("LETTA_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("LETTA_API_KEY environment variable not set")
	}

	if baseURL == "" {
		baseURL = "https://api.letta.com"
	}

	return &Client{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}, nil
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// ListBlocks lists all blocks matching the description search
func (c *Client) ListBlocks(descriptionSearch string) ([]Block, error) {
	path := "/v1/blocks?"
	params := url.Values{}
	if descriptionSearch != "" {
		params.Set("description_search", descriptionSearch)
	}
	path += params.Encode()

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var blocks []Block
	if err := json.Unmarshal(respBody, &blocks); err != nil {
		return nil, fmt.Errorf("failed to parse blocks: %w", err)
	}

	return blocks, nil
}

// GetBlock retrieves a specific block by label and owner
func (c *Client) GetBlock(label, ownerID string) (*Block, error) {
	path := "/v1/blocks?"
	params := url.Values{}
	params.Set("label", label)
	params.Set("description_search", ownerID)
	path += params.Encode()

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var blocks []Block
	if err := json.Unmarshal(respBody, &blocks); err != nil {
		return nil, fmt.Errorf("failed to parse blocks: %w", err)
	}

	if len(blocks) == 0 {
		return nil, nil
	}

	return &blocks[0], nil
}

// CreateBlock creates a new block
func (c *Client) CreateBlock(label, value, description string) (*Block, error) {
	body := map[string]string{
		"label":       label,
		"value":       value,
		"description": description,
	}

	respBody, err := c.doRequest("POST", "/v1/blocks", body)
	if err != nil {
		return nil, err
	}

	var block Block
	if err := json.Unmarshal(respBody, &block); err != nil {
		return nil, fmt.Errorf("failed to parse block: %w", err)
	}

	return &block, nil
}

// UpdateBlock updates an existing block
func (c *Client) UpdateBlock(blockID, value string) (*Block, error) {
	body := map[string]string{
		"value": value,
	}

	respBody, err := c.doRequest("PATCH", "/v1/blocks/"+blockID, body)
	if err != nil {
		return nil, err
	}

	var block Block
	if err := json.Unmarshal(respBody, &block); err != nil {
		return nil, fmt.Errorf("failed to parse block: %w", err)
	}

	return &block, nil
}

// DeleteBlock deletes a block
func (c *Client) DeleteBlock(blockID string) error {
	_, err := c.doRequest("DELETE", "/v1/blocks/"+blockID, nil)
	return err
}

// Agent represents a Letta agent
type Agent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

// ListAgents lists all agents, optionally filtered by name search
func (c *Client) ListAgents(nameSearch string) ([]Agent, error) {
	path := "/v1/agents/"
	params := url.Values{}
	if nameSearch != "" {
		params.Set("query_text", nameSearch)
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var agents []Agent
	if err := json.Unmarshal(respBody, &agents); err != nil {
		return nil, fmt.Errorf("failed to parse agents: %w", err)
	}

	return agents, nil
}
