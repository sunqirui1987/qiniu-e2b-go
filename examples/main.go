package main

import (
	"context"
	"fmt"
	"log"
	"os"

	e2b "github.com/sunqirui1987/qiniu-e2b-go"
)

func main() {
	// Set environment variables (or they can be set externally)
	os.Setenv("E2B_API_KEY", os.Getenv("E2B_API_KEY"))
	os.Setenv("E2B_API_URL", "https://cn-yangzhou-1-sandbox.qiniuapi.com")

	ctx := context.Background()

	// Create a new sandbox (similar to JS: const sbx = await Sandbox.create())
	sbx, err := e2b.Create(ctx, &e2b.SandboxOpts{
		Template:  "code-interpreter-v1", // Template ID
		TimeoutMs: 300000,                // 5 minutes
		EnvVars: map[string]string{
			"FOO": "bar",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}

	fmt.Printf("Sandbox created: %s\n", sbx.SandboxID())

	// Run Python code in the sandbox
	fmt.Println("\n=== Running Python code ===")
	pythonExecution, err := sbx.RunCode(`print("hello world from python")`, &e2b.RunCodeOpts{
		Language: e2b.Python,
	})
	if err != nil {
		log.Fatalf("Failed to run Python code: %v", err)
	}
	printExecution(pythonExecution)

	// Run JavaScript code in the sandbox
	fmt.Println("\n=== Running JavaScript code ===")
	jsExecution, err := sbx.RunCode(`console.log("hello world from javascript")`, &e2b.RunCodeOpts{
		Language: e2b.JavaScript,
	})
	if err != nil {
		log.Fatalf("Failed to run JavaScript code: %v", err)
	}
	printExecution(jsExecution)

	// Run more JavaScript with calculations
	fmt.Println("\n=== Running JavaScript with calculations ===")
	jsExecution2, err := sbx.RunCode(`
const sum = (a, b) => a + b;
console.log("1 + 2 =", sum(1, 2));
for (let i = 0; i < 3; i++) {
  console.log("iteration", i);
}
	`, &e2b.RunCodeOpts{
		Language: e2b.JavaScript,
	})
	if err != nil {
		log.Fatalf("Failed to run JavaScript code: %v", err)
	}
	printExecution(jsExecution2)

	// Kill the sandbox (similar to JS: await sbx.kill())
	err = sbx.Kill()
	if err != nil {
		log.Fatalf("Failed to kill sandbox: %v", err)
	}

	fmt.Println("\nSandbox killed successfully")
}

func printExecution(execution *e2b.Execution) {
	fmt.Printf("  Execution Count: %d\n", execution.ExecutionCount)
	if execution.Error != nil {
		fmt.Printf("  Error: %s - %s\n", execution.Error.Name, execution.Error.Value)
	}
	// Print stdout/stderr logs
	for _, logEntry := range execution.Logs {
		streamType := "stdout"
		if logEntry.IsError {
			streamType = "stderr"
		}
		fmt.Printf("  [%s] %s", streamType, logEntry.Line)
	}
	// Print results
	for _, result := range execution.Results {
		if result.Text != "" {
			fmt.Printf("  Result: %s\n", result.Text)
		}
		if result.HTML != "" {
			fmt.Printf("  HTML: %s\n", result.HTML)
		}
		if result.Markdown != "" {
			fmt.Printf("  Markdown: %s\n", result.Markdown)
		}
	}
}
