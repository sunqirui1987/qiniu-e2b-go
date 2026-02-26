package e2b

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type (
	// SandboxTemplate is a sandbox template.
	SandboxTemplate string

	// Sandbox is a code sandbox.
	//
	// The sandbox is like an isolated, but interactive system.
	Sandbox struct {
		ID       string                  `json:"sandboxID"`  // ID of the sandbox.
		ClientID string                  `json:"clientID"`   // ClientID of the sandbox.
		Cwd      string                  `json:"cwd"`        // Cwd is the sandbox's current working directory.
		apiKey   string                  `json:"-"`          // apiKey is the sandbox's api key.
		Template SandboxTemplate         `json:"templateID"` // Template of the sandbox.
		baseURL  string                  `json:"-"`          // baseAPIURL is the base api url of the sandbox.
		Metadata map[string]string       `json:"metadata"`   // Metadata of the sandbox.
		logger   *slog.Logger            `json:"-"`          // logger is the sandbox's logger.
		client   *http.Client            `json:"-"`          // client is the sandbox's http client.
		ws       *websocket.Conn         `json:"-"`          // ws is the sandbox's websocket connection.
		wsURL    func(s *Sandbox) string `json:"-"`          // wsURL is the sandbox's websocket url.
		Map      *sync.Map               `json:"-"`          // Map is the map of the sandbox.
		idCh     chan int                `json:"-"`          // idCh is the channel to generate ids for requests.
	}

	// Option is an option for the sandbox.
	Option func(*Sandbox)
)

const (
	// QiniuSandboxBaseURL is the base URL for Qiniu Sandbox.
	// Region: cn-yangzhou-1
	QiniuSandboxBaseURL       = "https://cn-yangzhou-1-sandbox.qiniuapi.com"
	defaultBaseURL            = QiniuSandboxBaseURL
	defaultWSScheme           = "wss"
	wsRoute                   = "/ws"
	fileRoute                 = "/file"
	sandboxesRoute            = "/sandboxes"  // (GET/POST /sandboxes)
	deleteSandboxRoute        = "/sandboxes/" // (DELETE /sandboxes/:id)
	notebookExecCell   Method = "notebook_execCell"
)

// NewSandbox creates a new sandbox.
func NewSandbox(
	ctx context.Context,
	apiKey string,
	opts ...Option,
) (*Sandbox, error) {
	sb := Sandbox{
		apiKey:   apiKey,
		Template: "base",
		baseURL:  defaultBaseURL,
		Metadata: map[string]string{
			"sdk": "e2b-go v1",
		},
		client: http.DefaultClient,
		logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
		idCh:   make(chan int),
		Map:    new(sync.Map),
		wsURL: func(s *Sandbox) string {
			return fmt.Sprintf("wss://49982-%s-%s.e2b.dev/ws", s.ID, s.ClientID)
		},
	}
	for _, opt := range opts {
		opt(&sb)
	}
	req, err := sb.newRequest(ctx, http.MethodPost, fmt.Sprintf("%s%s", sb.baseURL, sandboxesRoute), &sb)
	if err != nil {
		return &sb, err
	}
	err = sb.sendRequest(req, &sb)
	if err != nil {
		return &sb, err
	}
	var resp *http.Response
	sb.ws, resp, err = websocket.DefaultDialer.Dial(sb.wsURL(&sb), nil)
	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}
	if err != nil {
		return &sb, err
	}
	go sb.identify(ctx)
	go func() {
		err := sb.read(ctx)
		if err != nil {
			sb.logger.Error("failed to read sandbox", "error", err)
		}
	}()
	return &sb, nil
}

// ConnectSandbox connects to an existing sandbox.
func ConnectSandbox(
	ctx context.Context,
	sandboxID string,
	apiKey string,
	opts ...Option,
) (*Sandbox, error) {
	sb := Sandbox{
		ID:      sandboxID,
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		Metadata: map[string]string{
			"sdk": "e2b-go v1",
		},
		client: http.DefaultClient,
		logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
		idCh:   make(chan int),
		Map:    new(sync.Map),
		wsURL: func(s *Sandbox) string {
			return fmt.Sprintf("wss://49982-%s-%s.e2b.dev/ws", s.ID, s.ClientID)
		},
	}
	for _, opt := range opts {
		opt(&sb)
	}

	req, err := sb.newRequest(ctx, http.MethodGet, fmt.Sprintf("%s%s/%s", sb.baseURL, sandboxesRoute, sandboxID), nil)
	if err != nil {
		return &sb, err
	}
	err = sb.sendRequest(req, &sb)
	if err != nil {
		return &sb, err
	}

	var resp *http.Response
	sb.ws, resp, err = websocket.DefaultDialer.Dial(sb.wsURL(&sb), nil)
	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}
	if err != nil {
		return &sb, err
	}
	go sb.identify(ctx)
	go func() {
		err := sb.read(ctx)
		if err != nil {
			sb.logger.Error("failed to read sandbox", "error", err)
		}
	}()
	return &sb, nil
}

// KeepAlive keeps the sandbox alive.
func (s *Sandbox) KeepAlive(ctx context.Context, timeout time.Duration) error {
	body := struct {
		Duration int `json:"duration"`
	}{Duration: int(timeout.Seconds())}
	req, err := s.newRequest(ctx, http.MethodPost, fmt.Sprintf("%s/sandboxes/%s/refreshes", s.baseURL, s.ID), body)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("request to keep alive sandbox failed: %s", resp.Status)
	}
	return nil
}

// Reconnect reconnects to the sandbox.
func (s *Sandbox) Reconnect(ctx context.Context) (err error) {
	if err := s.ws.Close(); err != nil {
		return err
	}
	urlu := s.wsURL(s)
	var resp *http.Response
	s.ws, resp, err = websocket.DefaultDialer.Dial(urlu, nil)
	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}
	if err != nil {
		return err
	}
	go func() {
		err := s.read(ctx)
		if err != nil {
			fmt.Println(err)
		}
	}()
	return err
}

// Stop stops the sandbox.
func (s *Sandbox) Stop(ctx context.Context) error {
	req, err := s.newRequest(ctx, http.MethodDelete, fmt.Sprintf("%s%s%s", s.baseURL, deleteSandboxRoute, s.ID), nil)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("request to delete sandbox failed: %s", resp.Status)
	}
	return nil
}

// Close is an alias for Stop.
func (s *Sandbox) Close(ctx context.Context) error {
	return s.Stop(ctx)
}

// GetHost returns the host address for the specified port.
func (s *Sandbox) GetHost(port int) string {
	return fmt.Sprintf("%d-%s-%s.sandbox.qiniuapi.com", port, s.ID, s.ClientID)
}
