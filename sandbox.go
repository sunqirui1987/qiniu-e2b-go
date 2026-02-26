package e2b

import (
	"context"
	"fmt"
	"os"
	"time"
)

// Sandbox represents a code interpreter sandbox
// Similar to JS SDK's Sandbox class
type Sandbox struct {
	sandboxID   string
	templateID  string
	apiKey      string
	client      *Client
	localMode   bool

	// Files is the module for interacting with the sandbox filesystem
	// Similar to JS SDK's sandbox.files
	Files *Filesystem

	// Execution count for code execution
	executionCount int
}

// SandboxOpts represents options for creating a new Sandbox
type SandboxOpts struct {
	// Template ID for the sandbox
	Template string

	// API key for authentication
	APIKey string

	// Environment variables for the sandbox
	EnvVars map[string]string

	// Metadata for the sandbox
	Metadata map[string]string

	// Timeout for the sandbox in milliseconds (default: 5 minutes)
	TimeoutMs int
}

// DefaultSandboxOpts returns default options for creating a sandbox
func DefaultSandboxOpts() *SandboxOpts {
	return &SandboxOpts{
		Template:  "base",
		TimeoutMs: 300000, // 5 minutes
		EnvVars:   make(map[string]string),
		Metadata:  make(map[string]string),
	}
}

// Create creates a new sandbox from the default template
// This is the equivalent of Sandbox.create() in JS
func Create(ctx context.Context, opts ...*SandboxOpts) (*Sandbox, error) {
	var opt *SandboxOpts
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	} else {
		opt = DefaultSandboxOpts()
	}
	return NewSandbox(ctx, opt)
}

// NewSandbox creates a new sandbox with the specified options
func NewSandbox(ctx context.Context, opts *SandboxOpts) (*Sandbox, error) {
	if opts == nil {
		opts = DefaultSandboxOpts()
	}

	// Determine API key
	apiKey := opts.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("E2B_API_KEY")
	}

	// Check if we should use local mode
	localMode := isLocalMode() || apiKey == ""

	sbx := &Sandbox{
		apiKey:         apiKey,
		localMode:      localMode,
		executionCount: 0,
	}

	// Create client
	sbx.client = NewClient(apiKey)

	// Create sandbox via API or use local mode
	if sbx.localMode {
		// Local mode - use mock implementation
		sbx.sandboxID = generateSandboxID()
		sbx.templateID = opts.Template
	} else {
		// Remote mode - create sandbox via API
		req := &CreateSandboxRequest{
			TemplateID: opts.Template,
			EnvVars:    opts.EnvVars,
			Metadata:   opts.Metadata,
		}

		if opts.TimeoutMs > 0 {
			req.TimeoutMs = opts.TimeoutMs
		}

		resp, err := sbx.client.CreateSandbox(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to create sandbox: %w", err)
		}
		sbx.sandboxID = resp.SandboxID
		sbx.templateID = resp.TemplateID
	}

	// Initialize Filesystem
	sbx.Files = NewFilesystem(sbx.client, sbx.sandboxID, sbx.localMode)

	return sbx, nil
}

// isLocalMode checks if running in local mode
func isLocalMode() bool {
	return os.Getenv("E2B_LOCAL_MODE") == "true"
}

// generateSandboxID generates a unique sandbox ID for local mode
func generateSandboxID() string {
	return fmt.Sprintf("local-%d", time.Now().UnixNano())
}

// Kill terminates the sandbox
func (sbx *Sandbox) Kill() error {
	if sbx.localMode {
		// Local mode - just clear state
		return nil
	}

	// Remote mode - kill sandbox via API
	return sbx.client.KillSandbox(context.Background(), sbx.sandboxID)
}

// SandboxID returns the sandbox ID
func (sbx *Sandbox) SandboxID() string {
	return sbx.sandboxID
}

// TemplateID returns the template ID
func (sbx *Sandbox) TemplateID() string {
	return sbx.templateID
}

// RunCode executes Python code in the sandbox
// This is a convenience method for the most common use case
func (sbx *Sandbox) RunCode(code string, opts ...*RunCodeOpts) (*Execution, error) {
	ctx := context.Background()

	var opt *RunCodeOpts
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	} else {
		opt = DefaultRunCodeOpts()
	}

	return sbx.RunCodeWithContext(ctx, code, opt)
}

// RunCodeWithContext executes code in the sandbox with context
func (sbx *Sandbox) RunCodeWithContext(ctx context.Context, code string, opts *RunCodeOpts) (*Execution, error) {
	if opts == nil {
		opts = DefaultRunCodeOpts()
	}

	// Determine language
	language := opts.Language
	if language == "" {
		language = Python
	}

	// Build execution request
	request := &RunCodeRequest{
		Code:      code,
		Language:  string(language),
		TimeoutMs: opts.TimeoutMs,
		ContextID: opts.ContextID,
		EnvVars:   opts.EnvVars,
	}

	sbx.executionCount++

	if sbx.localMode {
		// Local mode - use mock implementation
		return sbx.runCodeLocal(code, opts)
	}

	// Remote mode - execute via API
	execution, err := sbx.client.RunCode(ctx, sbx.sandboxID, request)
	if err != nil {
		return nil, err
	}

	// Process callbacks
	if opts.OnStdout != nil && execution != nil {
		for _, log := range execution.Logs {
			opts.OnStdout(&OutputMessage{
				Line:      log.Line,
				Timestamp: log.Timestamp,
				Error:     log.IsError,
			})
		}
	}

	if opts.OnError != nil && execution != nil && execution.Error != nil {
		opts.OnError(execution.Error)
	}

	return execution, nil
}

// runCodeLocal executes code in local mode (mock implementation)
func (sbx *Sandbox) runCodeLocal(code string, opts *RunCodeOpts) (*Execution, error) {
	// Create a simple result
	result := &Result{
		Text:         fmt.Sprintf("Executed: %s", code),
		IsMainResult: true,
	}

	return &Execution{
		Results:        []*Result{result},
		Logs:           Logs{},
		Error:          nil,
		ExecutionCount: sbx.executionCount,
	}, nil
}
