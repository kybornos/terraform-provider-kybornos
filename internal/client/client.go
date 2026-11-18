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
	ID   string `json:"id"`
	Name string `json:"name"`
	Size string `json:"size"`
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
func (c *Client) Create(ctx context.Context, name, size string) (VM, error) {
	var result DeployResponse
	if err := c.request(ctx, http.MethodPost, "/v1/vms/deploy", map[string]string{
		"name": name,
		"size": size,
	}, &result); err != nil {
		return VM{}, err
	}
	if result.Status != "ok" {
		return VM{}, fmt.Errorf("deploy returned unexpected status %q", result.Status)
	}
	if result.Data == "" {
		return VM{}, errors.New("deploy response missing VM ID in data")
	}
	return VM{ID: result.Data, Name: name, Size: size}, nil
}

func (c *Client) Get(ctx context.Context, id string) (VM, error) {
	// The exact GET response shape is not yet confirmed. Accept either a
	// direct VM object or an {"status":"ok","data":{...}} envelope.
	var result struct {
		ID     string          `json:"id"`
		Name   string          `json:"name"`
		Size   string          `json:"size"`
		Status string          `json:"status"`
		Data   json.RawMessage `json:"data"`
	}
	if err := c.request(ctx, http.MethodGet, "/v1/vms/"+url.PathEscape(id), nil, &result); err != nil {
		return VM{}, err
	}
	vm := VM{ID: result.ID, Name: result.Name, Size: result.Size}
	if len(result.Data) != 0 && string(result.Data) != "null" {
		if result.Status != "" && result.Status != "ok" {
			return VM{}, fmt.Errorf("get VM returned unexpected status %q", result.Status)
		}
		if err := json.Unmarshal(result.Data, &vm); err != nil {
			return VM{}, fmt.Errorf("decode VM data: %w", err)
		}
	}
	if vm.ID == "" {
		vm.ID = id
	}
	return vm, nil
}

func (c *Client) Delete(ctx context.Context, id string) error {
	return c.request(ctx, http.MethodPost, "/v1/vms/"+url.PathEscape(id)+"/destroy", nil, nil)
}
