package e2b

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the default Qiniu Cloud Sandbox API base URL
	DefaultBaseURL = "https://cn-yangzhou-1-sandbox.qiniuapi.com"

	// DefaultTimeout is the default timeout for HTTP requests
	DefaultTimeout = 60 * time.Second

	// JupyterPort is the default Jupyter port (for code execution)
	JupyterPort = 49999
	// EnvdPort is the default envd port (for filesystem and commands)
	EnvdPort = 49983
)

// Client represents a Qiniu Cloud Sandbox API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *log.Logger
}

// NewClient creates a new Qiniu Cloud Sandbox API client
func NewClient(apiKey string) *Client {
	if apiKey == "" {
		apiKey = os.Getenv("E2B_API_KEY")
	}

	baseURL := os.Getenv("E2B_API_URL")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		logger: log.New(os.Stderr, "[E2B] ", log.LstdFlags),
	}
}

// SetBaseURL sets the API base URL
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// SetTimeout sets the HTTP client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// GetSandboxHost returns the host for a sandbox
func (c *Client) GetSandboxHost(sandboxID string, port int) string {
	// Format: {port}-{sandboxID}.{domain}
	// Extract domain from baseURL
	domain := "sandbox.qiniuapi.com"
	if c.baseURL == "https://cn-yangzhou-1-sandbox.qiniuapi.com" {
		domain = "cn-yangzhou-1.sandbox.qibox.com"
	}
	return fmt.Sprintf("%d-%s.%s", port, sandboxID, domain)
}

// doRequest performs an HTTP request to the API
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, response interface{}) error {
	return c.doRequestToURL(ctx, method, c.baseURL+path, body, response)
}

// doRequestToURL performs an HTTP request to a specific URL
func (c *Client) doRequestToURL(ctx context.Context, method, url string, body interface{}, response interface{}) error {
	var reqBody io.Reader
	var bodyStr string

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyStr = string(jsonBody)
		reqBody = bytes.NewReader(jsonBody)
	}

	// Debug logging
	c.logger.Printf("Request: %s %s", method, url)
	if c.apiKey != "" {
		c.logger.Printf("API Key: %s...", c.apiKey[:min(10, len(c.apiKey))])
	}
	if bodyStr != "" {
		c.logger.Printf("Body: %s", bodyStr)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Set authorization header
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	c.logger.Printf("Response Status: %d", resp.StatusCode)
	c.logger.Printf("Response Body: %s", string(respBody))

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if response != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, response); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CreateSandboxRequest represents the request to create a sandbox
type CreateSandboxRequest struct {
	TemplateID string            `json:"templateID,omitempty"`
	EnvVars    map[string]string `json:"envVars,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	TimeoutMs  int               `json:"timeoutMs,omitempty"`
}

// CreateSandboxResponse represents the response from creating a sandbox
type CreateSandboxResponse struct {
	SandboxID          string `json:"sandboxID"`
	TemplateID         string `json:"templateID,omitempty"`
	ClientID           string `json:"clientID,omitempty"`
	SandboxURL         string `json:"sandboxURL,omitempty"`
	Domain             string `json:"domain,omitempty"`
	EnvdAccessToken    string `json:"envdAccessToken,omitempty"`
	TrafficAccessToken string `json:"trafficAccessToken,omitempty"`
}

// CreateSandbox creates a new sandbox
func (c *Client) CreateSandbox(ctx context.Context, req *CreateSandboxRequest) (*CreateSandboxResponse, error) {
	var resp CreateSandboxResponse
	err := c.doRequest(ctx, "POST", "/sandboxes", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// KillSandbox kills a sandbox
func (c *Client) KillSandbox(ctx context.Context, sandboxID string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/sandboxes/%s", sandboxID), nil, nil)
}

// RunCodeRequest represents the request to run code
type RunCodeRequest struct {
	Code      string            `json:"code"`
	Language  string            `json:"language,omitempty"`
	TimeoutMs int               `json:"timeoutMs,omitempty"`
	ContextID string            `json:"context_id,omitempty"`
	EnvVars   map[string]string `json:"env_vars,omitempty"`
}

// RunCodeResponse represents the response from running code
type RunCodeResponse struct {
	Execution *Execution `json:"execution"`
}

// RunCode executes code in a sandbox via the sandbox host (not API)
// This calls the Jupyter endpoint on the sandbox directly
func (c *Client) RunCode(ctx context.Context, sandboxID string, req *RunCodeRequest) (*Execution, error) {
	// Build the sandbox host URL for Jupyter
	host := c.GetSandboxHost(sandboxID, JupyterPort)
	url := fmt.Sprintf("https://%s/execute", host)

	// Create request body
	body := map[string]interface{}{
		"code":       req.Code,
		"language":   req.Language,
		"context_id": req.ContextID,
		"env_vars":   req.EnvVars,
	}

	// Do request but handle streaming response manually
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	c.logger.Printf("Request: POST %s", url)
	if c.apiKey != "" {
		c.logger.Printf("API Key: %s...", c.apiKey[:min(10, len(c.apiKey))])
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse streaming NDJSON response
	return c.parseNDJSONResponse(resp.Body)
}

// parseNDJSONResponse parses NDJSON response line by line
func (c *Client) parseNDJSONResponse(body io.Reader) (*Execution, error) {
	execution := &Execution{
		Results: make([]*Result, 0),
		Logs:    make(Logs, 0),
	}

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		c.logger.Printf("Response Line: %s", line)

		var msg map[string]interface{}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue // Skip invalid lines
		}

		c.processExecutionMessage(msg, execution)
	}

	return execution, scanner.Err()
}

// processExecutionMessage processes a single execution message
func (c *Client) processExecutionMessage(msg map[string]interface{}, execution *Execution) {
	msgType, _ := msg["type"].(string)

	switch msgType {
	case "stdout":
		text, _ := msg["text"].(string)
		execution.Logs = append(execution.Logs, &Log{
			Line:      text,
			Timestamp: time.Now().UnixNano(),
			IsError:   false,
		})
	case "stderr":
		text, _ := msg["text"].(string)
		execution.Logs = append(execution.Logs, &Log{
			Line:      text,
			Timestamp: time.Now().UnixNano(),
			IsError:   true,
		})
	case "result":
		result := &Result{
			Text:         getString(msg, "text"),
			HTML:         getString(msg, "html"),
			Markdown:     getString(msg, "markdown"),
			IsMainResult: getBool(msg, "is_main_result"),
		}
		execution.Results = append(execution.Results, result)
	case "error":
		execution.Error = &ExecutionError{
			Name:      getString(msg, "name"),
			Value:     getString(msg, "value"),
			Traceback: getString(msg, "traceback"),
		}
	case "number_of_executions":
		if count, ok := msg["execution_count"].(float64); ok {
			execution.ExecutionCount = int(count)
		}
	}
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

// FilesystemRequest represents a filesystem request
type FilesystemRequest struct {
	SandboxID string
	Host      string
}

// ReadFile reads a file from the sandbox via the sandbox host
// Note: The envd API returns raw file content, not JSON
func (c *Client) ReadFile(ctx context.Context, sandboxID, path string) ([]byte, error) {
	host := c.GetSandboxHost(sandboxID, EnvdPort)
	url := fmt.Sprintf("https://%s/files?path=%s", host, path)

	// Do raw request to get file content
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// WriteFile writes a file to the sandbox via the sandbox host
// Note: The envd API expects POST /files with multipart/form-data
func (c *Client) WriteFile(ctx context.Context, sandboxID string, req *WriteFileRequest) error {
	host := c.GetSandboxHost(sandboxID, EnvdPort)
	url := fmt.Sprintf("https://%s/files?path=%s", host, req.Path)

	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Get filename from path
	filename := req.Path
	if idx := strings.LastIndex(req.Path, "/"); idx >= 0 {
		filename = req.Path[idx+1:]
	}

	// Add file field
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = part.Write(req.Data)
	if err != nil {
		return fmt.Errorf("failed to write file data: %w", err)
	}

	writer.Close()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	c.logger.Printf("Request: POST %s (multipart/form-data)", url)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.logger.Printf("Response Status: %d, Body: %s", resp.StatusCode, string(respBody))

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// ListFiles lists files in a directory via the sandbox host
// Note: Using the envd API /files endpoint
func (c *Client) ListFiles(ctx context.Context, sandboxID, path string) ([]*File, error) {
	host := c.GetSandboxHost(sandboxID, EnvdPort)
	// Try different possible endpoints
	urls := []string{
		fmt.Sprintf("https://%s/files/list?path=%s", host, path),
		fmt.Sprintf("https://%s/files?path=%s&list=true", host, path),
		fmt.Sprintf("https://%s/directories?path=%s", host, path),
	}

	var lastErr error
	for _, url := range urls {
		c.logger.Printf("Trying URL: %s", url)
		var resp struct {
			Entries []*File `json:"entries"`
			Files   []*File `json:"files"`
		}
		err := c.doRequestToURL(ctx, "GET", url, nil, &resp)
		if err == nil {
			if len(resp.Entries) > 0 {
				return resp.Entries, nil
			}
			return resp.Files, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

// RemoveFile removes a file from the sandbox via the sandbox host
func (c *Client) RemoveFile(ctx context.Context, sandboxID, path string) error {
	host := c.GetSandboxHost(sandboxID, EnvdPort)
	url := fmt.Sprintf("https://%s/files?path=%s", host, path)

	return c.doRequestToURL(ctx, "DELETE", url, nil, nil)
}

// MakeDir creates a directory in the sandbox via the sandbox host
func (c *Client) MakeDir(ctx context.Context, sandboxID string, req *MakeDirRequest) error {
	host := c.GetSandboxHost(sandboxID, EnvdPort)
	url := fmt.Sprintf("https://%s/files/mkdir", host)

	return c.doRequestToURL(ctx, "POST", url, req, nil)
}

// WriteFileRequest represents the request to write a file
type WriteFileRequest struct {
	Path string `json:"path"`
	Data []byte `json:"data"`
	Mode string `json:"mode,omitempty"`
}

// File represents a file in the sandbox
type File struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"`
	Mode  int    `json:"mode"`
}

// MakeDirRequest represents the request to create a directory
type MakeDirRequest struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive"`
	Mode      string `json:"mode,omitempty"`
}

// ContextRequest represents a context request
type ContextRequest struct {
	SandboxID string
}

// CreateContext creates a new code context
func (c *Client) CreateContext(ctx context.Context, sandboxID string, req *CreateContextRequest) (*Context, error) {
	host := c.GetSandboxHost(sandboxID, JupyterPort)
	url := fmt.Sprintf("https://%s/contexts", host)

	var resp Context
	err := c.doRequestToURL(ctx, "POST", url, req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateContextRequest represents the request to create a code context
type CreateContextRequest struct {
	Language string `json:"language"`
	Cwd      string `json:"cwd"`
}

// RemoveContext removes a code context
func (c *Client) RemoveContext(ctx context.Context, sandboxID, contextID string) error {
	host := c.GetSandboxHost(sandboxID, JupyterPort)
	url := fmt.Sprintf("https://%s/contexts/%s", host, contextID)

	return c.doRequestToURL(ctx, "DELETE", url, nil, nil)
}

// ListContexts lists all code contexts
func (c *Client) ListContexts(ctx context.Context, sandboxID string) ([]*Context, error) {
	host := c.GetSandboxHost(sandboxID, JupyterPort)
	url := fmt.Sprintf("https://%s/contexts", host)

	var resp struct {
		Contexts []*Context `json:"contexts"`
	}
	err := c.doRequestToURL(ctx, "GET", url, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Contexts, nil
}

// RestartContext restarts a code context
func (c *Client) RestartContext(ctx context.Context, sandboxID, contextID string) error {
	host := c.GetSandboxHost(sandboxID, JupyterPort)
	url := fmt.Sprintf("https://%s/contexts/%s/restart", host, contextID)

	return c.doRequestToURL(ctx, "POST", url, nil, nil)
}

// GetAccessToken gets an access token for the sandbox
func (c *Client) GetAccessToken(ctx context.Context, sandboxID string) (string, error) {
	var resp struct {
		AccessToken string `json:"accessToken"`
	}
	err := c.doRequest(ctx, "GET", fmt.Sprintf("/sandboxes/%s/token", sandboxID), nil, &resp)
	if err != nil {
		return "", err
	}
	return resp.AccessToken, nil
}