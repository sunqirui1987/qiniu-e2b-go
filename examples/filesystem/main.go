package main

import (
	"context"
	"fmt"
	"log"
	"os"

	e2b "github.com/qiniu/e2b-go"
)

func main() {
	// Set environment variables
	os.Setenv("E2B_API_KEY", os.Getenv("E2B_API_KEY"))
	os.Setenv("E2B_API_URL", "https://cn-yangzhou-1-sandbox.qiniuapi.com")

	ctx := context.Background()

	// Create a new sandbox
	sbx, err := e2b.Create(ctx, &e2b.SandboxOpts{
		Template:  "code-interpreter-v1",
		TimeoutMs: 300000,
	})
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	fmt.Printf("Sandbox created: %s\n", sbx.SandboxID())

	// Test 1: List files in /
	fmt.Println("\n=== Test 1: List files in / ===")
	files, err := sbx.Files.List("/")
	if err != nil {
		fmt.Printf("Failed to list files: %v\n", err)
	} else {
		fmt.Printf("Found %d entries:\n", len(files))
		for _, file := range files {
			fmt.Printf("  - %s (type: %v, size: %d)\n", file.Name, file.Type, file.Size)
		}
	}

	// Test 2: Write a file
	fmt.Println("\n=== Test 2: Write file ===")
	testContent := "Hello from Go SDK!"
	err = sbx.Files.WriteString("/home/user/test.txt", testContent)
	if err != nil {
		fmt.Printf("Failed to write file: %v\n", err)
	} else {
		fmt.Println("File written successfully")
	}

	// Test 3: Read the file
	fmt.Println("\n=== Test 3: Read file ===")
	content, err := sbx.Files.Read("/home/user/test.txt")
	if err != nil {
		fmt.Printf("Failed to read file: %v\n", err)
	} else {
		fmt.Printf("File content: %s\n", content)
	}

	// Test 4: Create directory
	fmt.Println("\n=== Test 4: Create directory ===")
	err = sbx.Files.MakeDir("/home/user/mydir")
	if err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
	} else {
		fmt.Println("Directory created successfully")
	}

	// Test 5: List files in /home/user
	fmt.Println("\n=== Test 5: List files in /home/user ===")
	files, err = sbx.Files.List("/home/user")
	if err != nil {
		fmt.Printf("Failed to list files: %v\n", err)
	} else {
		fmt.Printf("Found %d entries:\n", len(files))
		for _, file := range files {
			fmt.Printf("  - %s (type: %v, size: %d)\n", file.Name, file.Type, file.Size)
		}
	}

	// Test 6: Check if file exists
	fmt.Println("\n=== Test 6: Check file exists ===")
	exists, err := sbx.Files.Exists("/home/user/test.txt")
	if err != nil {
		fmt.Printf("Failed to check file: %v\n", err)
	} else {
		fmt.Printf("File exists: %v\n", exists)
	}

	// Test 7: Get file info
	fmt.Println("\n=== Test 7: Get file info ===")
	info, err := sbx.Files.GetInfo("/home/user/test.txt")
	if err != nil {
		fmt.Printf("Failed to get file info: %v\n", err)
	} else {
		fmt.Printf("File info: name=%s, type=%v, size=%d\n", info.Name, info.Type, info.Size)
	}

	// Kill the sandbox
	err = sbx.Kill()
	if err != nil {
		log.Fatalf("Failed to kill sandbox: %v", err)
	}
	fmt.Println("\nSandbox killed successfully")
}
