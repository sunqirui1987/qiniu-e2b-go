package e2b

// Language represents a programming language for code execution
type Language string

const (
	// Python is the Python programming language
	Python Language = "python"
	// JavaScript is the JavaScript programming language
	JavaScript Language = "javascript"
	// TypeScript is the TypeScript programming language
	TypeScript Language = "typescript"
	// Bash is the Bash shell
	Bash Language = "bash"
	// Go is the Go programming language
	GoLang Language = "go"
	// Rust is the Rust programming language
	Rust Language = "rust"
	// Java is the Java programming language
	Java Language = "java"
)

// RuntimeName returns the runtime name for a language
func RuntimeName(lang Language) string {
	switch lang {
	case Python:
		return "python3"
	case JavaScript:
		return "node"
	case TypeScript:
		return "ts-node"
	case Bash:
		return "bash"
	case GoLang:
		return "go"
	case Rust:
		return "rust"
	case Java:
		return "java"
	default:
		return string(lang)
	}
}
