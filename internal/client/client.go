package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

type VM struct {
	ID               string `json:"id"`
	Hostname         string
	VMName           string `json:"vm_name"`
	Size             VMSize `json:"size"`
	Public           bool
	Encrypt          string
	AutodiskTemplate bool
	Interfaces       []VMInterface
}

type VMSize struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	CPU    int64  `json:"cpu"`
	Memory int64  `json:"memory"`
	Disk   int64  `json:"disk"`
}

type VMInterface struct {
	IP4     string `json:"ip4"`
	LLAddr  string `json:"lladdr"`
	Primary bool   `json:"primary"`
	Public  bool   `json:"public"`
}

type VMResponse struct {
	Status string `json:"status"`
	Data   struct {
		ID       string `json:"id"`
		Hostname string `json:"hostname"`
		VMName   string `json:"vm_name"`
		State    string `json:"state"`
		Data     struct {
			Size       VMSize        `json:"size"`
			Public     bool          `json:"public"`
			Interfaces []VMInterface `json:"interfaces"`
		} `json:"data"`
	} `json:"data"`
}

type Error struct {
	Status  int
	Message string
}

type DeployResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func (e *Error) Error() string { return fmt.Sprintf("KYBORNOS API HTTP %d: %s", e.Status, e.Message) }

func New(endpoint, token string) (*Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return nil, fmt.Errorf("endpoint must be a valid http(s) URL")
	}
	if u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("endpoint must not contain credentials, query, or fragment")
	}
	if token == "" {
		return nil, errors.New("missing API key")
	}
	return &Client{BaseURL: strings.TrimRight(endpoint, "/"), Token: token, HTTP: &http.Client{Timeout: 30 * time.Second}}, nil
}
func (c *Client) request(ctx context.Context, method, path string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := http.StatusText(resp.StatusCode)
		var apiErr struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if json.Unmarshal(b, &apiErr) == nil {
			if apiErr.Message != "" {
				msg = apiErr.Message
			} else if apiErr.Error != "" {
				msg = apiErr.Error
			}
		}
		return &Error{Status: resp.StatusCode, Message: msg}
	}
	if out != nil {
		if len(b) == 0 {
			return errors.New("empty API response")
		}
		if err := json.Unmarshal(b, out); err != nil {
			return fmt.Errorf("decode API response: %w", err)
		}
	}
	return nil
}
func (c *Client) Create(ctx context.Context, hostname, size string, public bool, encrypt string, autodisk_template *bool, autoinstall map[string]any, srvTags []map[string]any) (VM, error) {
	var result DeployResponse

	payload := map[string]any{
		"name":   hostname,
		"size":   size,
		"public": public,
	}

	if len(autoinstall) > 0 {
		payload["autoinstall"] = autoinstall
	}
	if len(encrypt) > 0 {
		payload["encrypt"] = encrypt
	}
	if autodisk_template != nil {
		payload["autodisk_template"] = *autodisk_template
	}
	if srvTags != nil {
		payload["srv_tags"] = srvTags
	}
	if err := c.request(ctx, http.MethodPost, "/v1/vms/deploy", payload, &result); err != nil {
		return VM{}, err
	}

	if result.Status != "ok" {
		return VM{}, fmt.Errorf("deploy returned unexpected status %q", result.Status)
	}

	if result.Data == "" {
		return VM{}, errors.New("deploy response missing VM ID in data")
	}

	return VM{ID: result.Data}, nil
}

func (c *Client) Get(ctx context.Context, id string) (VM, error) {
	var result VMResponse
	if err := c.request(ctx, http.MethodGet, "/v1/vms/"+url.PathEscape(id), nil, &result); err != nil {
		return VM{}, err
	}
	vm := VM{
		ID:         result.Data.ID,
		Hostname:   result.Data.Hostname,
		VMName:     result.Data.VMName,
		Size:       result.Data.Data.Size,
		Public:     result.Data.Data.Public,
		Interfaces: result.Data.Data.Interfaces,
	}
	if vm.ID == "" {
		return VM{}, fmt.Errorf("VM response missing ID")
	}

	if vm.Size.Name == "" {
		return VM{}, fmt.Errorf("VM response missing size")
	}

	return vm, nil
}

func (c *Client) Delete(ctx context.Context, id string) error {
	return c.request(ctx, http.MethodPost, "/v1/vms/"+url.PathEscape(id)+"/destroy", nil, nil)
}
