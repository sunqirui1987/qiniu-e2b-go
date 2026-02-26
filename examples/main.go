package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sunqirui1987/qiniu-e2b-go"
)

func main() {
	ctx := context.Background()

	// Replace with your Qiniu API key
	// Get it from: https://portal.qiniu.com/
	apiKey := "your-qiniu-api-key"

	// Example 1: Basic Sandbox with process execution
	basicSandbox(ctx, apiKey)

	// Example 2: Code Interpreter
	// codeInterpreter(ctx, apiKey)

	// Example 3: File operations
	// fileOperations(ctx, apiKey)
}

func basicSandbox(ctx context.Context, apiKey string) {
	fmt.Println("=== Basic Sandbox Example ===")

	// Create a new sandbox
	sb, err := e2b.NewSandbox(ctx, apiKey)
	if err != nil {
		log.Fatal(err)
	}
	defer sb.Close(ctx)

	fmt.Printf("Sandbox ID: %s\n", sb.ID)

	// Run a shell command
	proc, err := sb.NewProcess("echo hello world")
	if err != nil {
		log.Fatal(err)
	}

	if err := proc.Start(ctx); err != nil {
		log.Fatal(err)
	}

	// Subscribe to stdout
	stdout, _ := proc.SubscribeStdout(ctx)
	for event := range stdout {
		fmt.Println("stdout:", event.Params.Result.Line)
	}

	// List files in the sandbox
	files, err := sb.Ls(ctx, "/")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Files in root:")
	for _, f := range files {
		fmt.Printf("  %s (dir: %v)\n", f.Name, f.IsDir)
	}
}

func codeInterpreter(ctx context.Context, apiKey string) {
	fmt.Println("=== Code Interpreter Example ===")

	// Create a new code interpreter sandbox
	sbx, err := e2b.NewCodeInterpreter(ctx, apiKey)
	if err != nil {
		log.Fatal(err)
	}
	defer sbx.Close(ctx)

	// Execute code - first cell
	_, err = sbx.RunCode(ctx, "x = 1")
	if err != nil {
		log.Fatal(err)
	}

	// Execute code - second cell (state is preserved)
	execution, err := sbx.RunCode(ctx, "x += 1; x")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Result:", execution.Text()) // Outputs: 2

	// More complex example with data visualization
	exec, err := sbx.RunCode(ctx, `
import matplotlib.pyplot as plt
import numpy as np

x = np.linspace(0, 10, 100)
y = np.sin(x)

plt.figure(figsize=(8, 4))
plt.plot(x, y)
plt.title('Sine Wave')
plt.savefig('sine.png')
print('Plot saved to sine.png')
`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(exec.Text())
}

func fileOperations(ctx context.Context, apiKey string) {
	fmt.Println("=== File Operations Example ===")

	sb, err := e2b.NewSandbox(ctx, apiKey)
	if err != nil {
		log.Fatal(err)
	}
	defer sb.Close(ctx)

	// Write a file
	err = sb.Write(ctx, "/home/user/test.txt", []byte("Hello from Qiniu Sandbox!"))
	if err != nil {
		log.Fatal(err)
	}

	// Read the file
	content, err := sb.Read(ctx, "/home/user/test.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("File content:", content)

	// List files
	files, err := sb.Ls(ctx, "/home/user")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Files in /home/user:")
	for _, f := range files {
		fmt.Printf("  %s (dir: %v)\n", f.Name, f.IsDir)
	}

	// Create a directory
	err = sb.Mkdir(ctx, "/home/user/newdir")
	if err != nil {
		log.Fatal(err)
	}
}
