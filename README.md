# qiniu-e2b-go

Go SDK for [Qiniu Sandbox](https://developer.qiniu.com/las/13283/sandbox-quickstart) - Cloud runtime for AI agents.

Inspired by the [official E2B Code Interpreter SDKs](https://github.com/e2b-dev/code-interpreter).

Qiniu Sandbox is an open-source infrastructure that allows you to run AI-generated code in secure isolated sandboxes in the cloud.
It's compatible with E2B API.

## Installation

```bash
go get github.com/sunqirui1987/qiniu-e2b-go
```

## Quick Start

### 1. Get your Qiniu API key

1. Sign up to Qiniu [here](https://portal.qiniu.com/).
2. Get your API key from the console.

### 2. Execute code with Code Interpreter

The `CodeInterpreter` provides a stateful Jupyter-like environment.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sunqirui1987/qiniu-e2b-go"
)

func main() {
	ctx := context.Background()

	// Create a new code interpreter sandbox
	// By default it uses the "code-interpreter-v1" template
	sbx, err := e2b.NewCodeInterpreter(ctx, "your-qiniu-api-key")
	if err != nil {
		log.Fatal(err)
	}
	defer sbx.Close(ctx)

	// Execute code
	_, err = sbx.RunCode(ctx, "x = 1")
	if err != nil {
		log.Fatal(err)
	}

	// State is preserved between calls
	execution, err := sbx.RunCode(ctx, "x += 1; x")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(execution.Text()) // Outputs: 2
}
```

### 3. Run shell commands

You can also run shell commands in the sandbox:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sunqirui1987/qiniu-e2b-go"
)

func main() {
	ctx := context.Background()

	sb, err := e2b.NewSandbox(ctx, "your-qiniu-api-key")
	if err != nil {
		log.Fatal(err)
	}
	defer sb.Close(ctx)

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
}
```

## Features

- **Code Interpreter**: Stateful execution of Python/JS code with rich output support (charts, images).
- **Sandbox Lifecycle**: Create, keep alive, reconnect, and stop sandboxes.
- **Filesystem Operations**: Read, write, list, mkdir, and watch for changes.
- **Process Execution**: Start processes with environment variables and working directory.
- **Event Streaming**: Subscribe to stdout, stderr, and exit events.

## Configuration

### Region

By default, the SDK uses `cn-yangzhou-1` region. You can change it using:

```go
sb, err := e2b.NewSandbox(ctx, apiKey, e2b.WithRegion("cn-yangzhou-1"))
```

### Custom Base URL

For advanced usage, you can set a custom base URL:

```go
sb, err := e2b.NewSandbox(ctx, apiKey, e2b.WithBaseURL("https://custom-sandbox.qiniuapi.com"))
```

## Documentation

- [Qiniu Sandbox Documentation](https://developer.qiniu.com/las/13283/sandbox-quickstart)

## License

MIT
