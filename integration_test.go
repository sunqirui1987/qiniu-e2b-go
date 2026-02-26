package e2b_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/sunqirui1987/qiniu-e2b-go"
	"github.com/stretchr/testify/assert"
)

func isIntegrationTest() bool {
	return os.Getenv("QINIU_API_KEY") != "" || os.Getenv("E2B_API_KEY") != ""
}

func getAPIKey(t *testing.T) string {
	apiKey := os.Getenv("QINIU_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("E2B_API_KEY")
	}
	if apiKey == "" {
		t.Skip("QINIU_API_KEY or E2B_API_KEY not set, skipping integration test")
	}
	return apiKey
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestPostSandbox(t *testing.T) {
	if !isIntegrationTest() {
		t.Skip("skipping integration test")
	}
	a := assert.New(t)
	ctx := context.Background()
	sb, err := e2b.NewSandbox(
		ctx,
		getAPIKey(t),
		e2b.WithLogger(testLogger()),
	)
	a.NoError(err, "NewSandbox error")
	lsr, err := sb.Ls(ctx, ".")
	a.NoError(err)
	for _, name := range []string{"boot", "code", "dev", "etc", "home"} {
		a.Contains(lsr, e2b.LsResult{
			Name:  name,
			IsDir: true,
		})
	}
	err = sb.Mkdir(ctx, "heelo")
	a.NoError(err)
	lsr, err = sb.Ls(ctx, "/")
	a.NoError(err)
	a.Contains(lsr, e2b.LsResult{
		Name:  "heelo",
		IsDir: true,
	})
}

func TestWriteRead(t *testing.T) {
	if !isIntegrationTest() {
		t.Skip("skipping integration test")
	}
	filePath := "test.txt"
	content := "Hello, world!"
	a := assert.New(t)
	ctx := context.Background()
	sb, err := e2b.NewSandbox(
		ctx,
		getAPIKey(t),
		e2b.WithLogger(testLogger()),
	)
	a.NoError(err, "NewSandbox error")
	err = sb.Write(ctx, filePath, []byte(content))
	a.NoError(err, "Write error")
	readContent, err := sb.Read(ctx, filePath)
	a.NoError(err, "Read error")
	a.Equal(content, readContent, "Read content does not match written content")
	readBytesContent, err := sb.ReadBytes(ctx, filePath)
	a.NoError(err, "ReadBytes error")
	a.Equal(content, string(readBytesContent), "ReadBytes content does not match written content")
	err = sb.Stop(ctx)
	a.NoError(err, "Stop error")
}

func TestCreateProcess(t *testing.T) {
	if !isIntegrationTest() {
		t.Skip("skipping integration test")
	}
	a := assert.New(t)
	ctx := context.Background()
	sb, err := e2b.NewSandbox(
		ctx,
		getAPIKey(t),
		e2b.WithLogger(testLogger()),
	)
	a.NoError(err, "NewSandbox error")
	proc, err := sb.NewProcess("echo 'Hello World!'", e2b.ProcessWithEnv(map[string]string{
		"FOO": "bar",
	}))
	a.NoError(err, "could not create process")
	err = proc.Start(ctx)
	a.NoError(err)
	proc, err = sb.NewProcess("sleep 2 && echo 'Hello World!'")
	a.NoError(err, "could not create process")
	err = proc.Start(ctx)
	a.NoError(err)
	stdOutEvents, errCh := proc.SubscribeStdout(ctx)
	select {
	case err := <-errCh:
		t.Fatalf("failed to subscribe to stdout: %v", err)
	case event := <-stdOutEvents:
		jsonBytes, err := json.MarshalIndent(&event, "", "  ")
		if err != nil {
			a.Error(err)
			return
		}
		t.Logf("test got event: %s", string(jsonBytes))
	}
}

func TestFilesystemSubscribe(t *testing.T) {
	if !isIntegrationTest() {
		t.Skip("skipping integration test")
	}
	a := assert.New(t)
	ctx := context.Background()
	sb, err := e2b.NewSandbox(
		ctx,
		getAPIKey(t),
		e2b.WithLogger(testLogger()),
		e2b.WithCwd("/tmp"),
	)
	a.NoError(err, "NewSandbox error")
	events := make(chan e2b.Event)
	err = sb.Watch(ctx, "/tmp/", events)
	a.NoError(err)
	go func() {
		for event := range events {
			jsonBytes, err := json.MarshalIndent(event, "", "  ")
			if err != nil {
				a.Error(err)
				return
			}
			t.Logf("test got event: %s", string(jsonBytes))
		}
	}()
	err = sb.Write(ctx, "/tmp/file.txt", []byte("Hello World!"))
	a.NoError(err)
	err = sb.Write(ctx, "/tmp/file2.txt", []byte("Hello World!"))
	a.NoError(err)
	time.Sleep(3 * time.Second)
}

func TestKeepAlive(t *testing.T) {
	if !isIntegrationTest() {
		t.Skip("skipping integration test")
	}
	a := assert.New(t)
	ctx := context.Background()
	sb, err := e2b.NewSandbox(
		ctx,
		getAPIKey(t),
		e2b.WithLogger(testLogger()),
	)
	a.NoError(err, "NewSandbox error")
	err = sb.KeepAlive(ctx, time.Minute*2)
	a.NoError(err)
}
