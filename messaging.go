package e2b

// OutputMessage represents an output message from the sandbox code execution
// Similar to JS SDK's OutputMessage
type OutputMessage struct {
	// The output line
	Line string

	// Unix epoch in nanoseconds
	Timestamp int64

	// Whether the output is an error
	Error bool

	// Execution count
	ExecutionCount int

	// Error object (if this is an error message)
	ExecutionErr *ExecutionError
}

// ExecutionError represents an error that occurred during the execution of a cell
// Similar to JS SDK's ExecutionError
type ExecutionError struct {
	// Name of the error
	Name string

	// Value of the error
	Value string

	// The raw traceback of the error
	Traceback string
}

// Result represents the data to be displayed as a result of executing a cell
// Similar to JS SDK's Result
type Result struct {
	// Text representation of the result
	Text string

	// HTML representation of the data
	HTML string

	// Markdown representation of the data
	Markdown string

	// SVG representation of the data
	SVG string

	// JSON representation of the data
	JSON string

	// Whether this is the main result
	IsMainResult bool

	// Chart data (if the result contains a chart)
	Chart any // Can be *LineChart, *ScatterChart, etc.
}

// Log represents a single log entry
type Log struct {
	Line      string
	Timestamp int64
	IsError   bool
}

// Logs represents a collection of log entries
type Logs []*Log

// Execution represents the result of executing a cell
// Similar to JS SDK's Execution
type Execution struct {
	// Results from the execution
	Results []*Result

	// Logs from stdout/stderr
	Logs Logs

	// Error that occurred during execution
	Error *ExecutionError

	// Execution count
	ExecutionCount int
}

// Context represents a code execution context
// Similar to JS SDK's Context
type Context struct {
	// The ID of the context
	ID string

	// The language of the context
	Language string

	// The working directory of the context
	Cwd string
}
