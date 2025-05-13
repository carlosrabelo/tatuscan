package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Client talks to the TatuScan API.
type Client struct {
	base   string
	client *http.Client
}

// New creates an API client.
func New(baseURL string) *Client {
	return &Client{
		base: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Machine is an inventory item from the API.
type Machine struct {
	MachineID          string  `json:"machine_id"`
	Hostname           string  `json:"hostname"`
	IP                 string  `json:"ip"`
	OS                 string  `json:"os"`
	OSVersion          *string `json:"os_version"`
	CPUPercent         float64 `json:"cpu_percent"`
	MemoryTotalMB      int64   `json:"memory_total_mb"`
	MemoryUsedMB       *int64  `json:"memory_used_mb"`
	ComputerModel      *string `json:"computer_model"`
	ComputerActivation *string `json:"computer_activation"`
	ActivationDays     *int    `json:"activation_days"`
	CreatedAt          *string `json:"created_at"`
	UpdatedAt          *string `json:"updated_at"`
}

// OSCount is OS distribution.
type OSCount struct {
	OS    string `json:"os"`
	Count int    `json:"count"`
}

// VersionCount is version distribution.
type VersionCount struct {
	Version string `json:"version"`
	Count   int    `json:"count"`
}

// AgeCount is age distribution.
type AgeCount struct {
	Range string `json:"range"`
	Count int    `json:"count"`
}

// OnlineStats is fleet online/offline counts.
type OnlineStats struct {
	Online  int    `json:"online"`
	Offline int    `json:"offline"`
	After   string `json:"after"`
}

// ListMachines GET /api/machines
func (c *Client) ListMachines(ctx context.Context) ([]Machine, error) {
	var resp struct {
		Items []Machine `json:"items"`
	}
	if err := c.get(ctx, "/api/machines", &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// StatsOS GET /api/stats/os
func (c *Client) StatsOS(ctx context.Context) ([]OSCount, error) {
	var resp struct {
		Items []OSCount `json:"items"`
	}
	if err := c.get(ctx, "/api/stats/os", &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// StatsVersions GET /api/stats/versions
func (c *Client) StatsVersions(ctx context.Context, top int) ([]VersionCount, error) {
	path := "/api/stats/versions?top=" + strconv.Itoa(top)
	var resp struct {
		Items []VersionCount `json:"items"`
	}
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// StatsAge GET /api/stats/age
func (c *Client) StatsAge(ctx context.Context) ([]AgeCount, error) {
	var resp struct {
		Items []AgeCount `json:"items"`
	}
	if err := c.get(ctx, "/api/stats/age", &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// StatsOnline GET /api/stats/online
func (c *Client) StatsOnline(ctx context.Context, after string) (OnlineStats, error) {
	path := "/api/stats/online"
	if after != "" {
		path += "?after=" + after
	}
	var resp OnlineStats
	if err := c.get(ctx, path, &resp); err != nil {
		return OnlineStats{}, err
	}
	return resp, nil
}

func (c *Client) get(ctx context.Context, path string, dest any) error {
	u := c.base + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("api %s: status %d", path, res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(dest)
}
