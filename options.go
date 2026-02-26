package e2b

// RunCodeOpts represents options for running code
type RunCodeOpts struct {
	// Callback for handling stdout messages
	OnStdout func(*OutputMessage)

	// Callback for handling stderr messages
	OnStderr func(*OutputMessage)

	// Callback for handling the final execution result
	OnResult func(*Result)

	// Callback for handling the ExecutionError object
	OnError func(*ExecutionError)

	// Custom environment variables for code execution
	EnvVars map[string]string

	// Timeout for the code execution in milliseconds
	TimeoutMs int

	// Language for code execution
	Language Language

	// Context ID for code execution
	ContextID string
}

// DefaultRunCodeOpts returns default options for running code
func DefaultRunCodeOpts() *RunCodeOpts {
	return &RunCodeOpts{
		EnvVars:   make(map[string]string),
		TimeoutMs: 60000, // 60 seconds
		Language:  Python,
	}
}

// CreateCodeContextOpts represents options for creating a code context
type CreateCodeContextOpts struct {
	// Working directory for the context
	Cwd string

	// Language for the context
	Language Language

	// Timeout for the request in milliseconds
	RequestTimeoutMs int
}

// DefaultCreateCodeContextOpts returns default options for creating a code context
func DefaultCreateCodeContextOpts() *CreateCodeContextOpts {
	return &CreateCodeContextOpts{
		Cwd:              "/home/user",
		Language:         Python,
		RequestTimeoutMs: 30000, // 30 seconds
	}
}
