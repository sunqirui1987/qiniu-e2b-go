// Package e2b provides a Go SDK for Qiniu Sandbox (https://developer.qiniu.com/las/13283/sandbox-quickstart),
// a cloud runtime for AI agents that provides secure sandboxed
// environments for code execution.
//
// This SDK is compatible with E2B API and uses Qiniu's sandbox service as the default backend.
//
// Usage:
//
//	package main
//
//	import (
//		"context"
//		"fmt"
//		"log"
//
//		"github.com/sunqirui1987/qiniu-e2b-go"
//	)
//
//	func main() {
//		ctx := context.Background()
//		apiKey := "your-qiniu-api-key" // Get from https://portal.qiniu.com/
//
//		// Create a new sandbox
//		sb, err := e2b.NewSandbox(ctx, apiKey)
//		if err != nil {
//			log.Fatal(err)
//		}
//		defer sb.Close(ctx)
//
//		// Run a command
//		proc, err := sb.NewProcess("echo hello")
//		if err != nil {
//			log.Fatal(err)
//		}
//		if err := proc.Start(ctx); err != nil {
//			log.Fatal(err)
//		}
//	}
package e2b
