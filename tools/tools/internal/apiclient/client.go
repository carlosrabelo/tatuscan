package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// DefaultAPIBase is used when neither TATUSCAN_URL nor --api-base is set.
	DefaultAPIBase = "http://localhost:8040/api"
	envURL         = "TATUSCAN_URL"
	envToken       = "TATUSCAN_API_TOKEN"
)

// Machine is an inventory item returned by the API.
type Machine struct {
	MachineID          string  `json:"machine_id"`
	Hostname           string  `json:"hostname"`
	IP                 string  `json:"ip"`
	OS                 string  `json:"os"`
	OSVersion          *string `json:"os_version"`
	CPUPercent         float64 `json:"cpu_percent"`
	MemoryTotalMB      int64   `json:"memory_total_mb"`
	MemoryUsedMB       *int64  `json:"memory_used_mb"`
	ComputerActivation *string `json:"computer_activation"`
	CreatedAt          *string `json:"created_at"`
	UpdatedAt          *string `json:"updated_at"`
}

// Client talks to the TatuScan API.
type Client struct {
	base   string
	token  string
	client *http.Client
}

// New creates an API client. token may be empty when the API has no auth.
func New(baseURL, token string) *Client {
	return &Client{
		base:  strings.TrimRight(baseURL, "/"),
		token: strings.TrimSpace(token),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func normalizeAPIBase(base string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		return DefaultAPIBase
	}
	if !strings.HasSuffix(base, "/api") {
		base += "/api"
	}
	return base
}

// ResolveAPIBase prefers TATUSCAN_URL, then flag, then DefaultAPIBase.
// A host without a trailing /api suffix gets /api appended.
func ResolveAPIBase(flagBase string) string {
	if env := strings.TrimSpace(os.Getenv(envURL)); env != "" {
		return normalizeAPIBase(env)
	}
	if strings.TrimSpace(flagBase) != "" {
		return normalizeAPIBase(flagBase)
	}
	return DefaultAPIBase
}

// ResolveToken returns TATUSCAN_API_TOKEN (optional Bearer for writes).
func ResolveToken() string {
	return strings.TrimSpace(os.Getenv(envToken))
}

// ListMachines GET /machines, falling back to /inventory on 404/405.
func (c *Client) ListMachines(ctx context.Context) ([]Machine, error) {
	var last error
	for _, path := range []string{"/machines", "/inventory"} {
		var resp struct {
			Items []Machine `json:"items"`
		}
		err := c.do(ctx, http.MethodGet, path, nil, &resp)
		if err == nil {
			if resp.Items == nil {
				return []Machine{}, nil
			}
			return resp.Items, nil
		}
		if status, ok := StatusOf(err); ok && (status == http.StatusNotFound || status == http.StatusMethodNotAllowed) {
			slog.Debug("endpoint unavailable, trying next", "path", path, "status", status)
			last = err
			continue
		}
		return nil, err
	}
	if last != nil {
		return nil, last
	}
	return []Machine{}, nil
}

// CreateMachine POST /machines.
func (c *Client) CreateMachine(ctx context.Context, payload map[string]any) (string, error) {
	var resp struct {
		Message string `json:"message"`
	}
	if err := c.do(ctx, http.MethodPost, "/machines", payload, &resp); err != nil {
		return "", err
	}
	if resp.Message == "" {
		return "OK", nil
	}
	return resp.Message, nil
}

// PatchActivation PATCH /machines/{id} with computer_activation.
func (c *Client) PatchActivation(ctx context.Context, machineID, isoDate string) error {
	return c.do(ctx, http.MethodPatch, "/machines/"+machineID, map[string]any{
		"computer_activation": isoDate,
	}, nil)
}

// DeleteMachine DELETE /machines/{id}.
func (c *Client) DeleteMachine(ctx context.Context, machineID string) error {
	return c.do(ctx, http.MethodDelete, "/machines/"+machineID, nil, nil)
}

func (c *Client) do(ctx context.Context, method, path string, body any, dest any) error {
	var rdr io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		rdr = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if res.StatusCode >= 300 {
		return &HTTPError{Status: res.StatusCode, Body: strings.TrimSpace(string(raw)), Path: path}
	}
	if dest == nil || len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

// HTTPError is a non-2xx API response.
type HTTPError struct {
	Status int
	Body   string
	Path   string
}

func (e *HTTPError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("api %s: status %d: %s", e.Path, e.Status, e.Body)
	}
	return fmt.Sprintf("api %s: status %d", e.Path, e.Status)
}

// StatusOf extracts an HTTP status from err when it is an *HTTPError.
func StatusOf(err error) (int, bool) {
	var he *HTTPError
	if errors.As(err, &he) {
		return he.Status, true
	}
	return 0, false
}
