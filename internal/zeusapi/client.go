// Copyright (c) WANIX Inc.
// SPDX-License-Identifier: MPL-2.0

package zeusapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("status %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("status %d", e.StatusCode)
}

func (e *APIError) NotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

func NewClient(baseURL, token string, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	return &Client{
		baseURL:    strings.TrimRight(parsed.String(), "/"),
		token:      token,
		httpClient: httpClient,
	}, nil
}

func (c *Client) do(ctx context.Context, method, path string, payload any, out any) error {
	fullURL := c.baseURL + path

	var body io.Reader
	if payload != nil {
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(payload); err != nil {
			return fmt.Errorf("encode payload: %w", err)
		}
		body = buf
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var eResp struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&eResp)
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    eResp.Error,
		}
	}

	if out == nil {
		return nil
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

type CreatePoolRequest struct {
	Start   int64  `json:"start"`
	Gateway int64  `json:"gateway"`
	Size    int64  `json:"size"`
	Region  string `json:"region"`
}

type CreatePoolResponse struct {
	ID string `json:"id"`
}

type PoolDetail struct {
	ID           string  `json:"id"`
	Region       string  `json:"region"`
	FriendlyName string  `json:"friendlyName"`
	Begin        string  `json:"begin"`
	End          string  `json:"end"`
	Gateway      string  `json:"gateway"`
	State        []int64 `json:"state"`
}

func (c *Client) CreatePool(ctx context.Context, req CreatePoolRequest) (CreatePoolResponse, error) {
	var resp CreatePoolResponse
	err := c.do(ctx, http.MethodPost, "/pools", req, &resp)
	return resp, err
}

func (c *Client) GetPoolByID(ctx context.Context, id string) (PoolDetail, error) {
	var resp PoolDetail
	err := c.do(ctx, http.MethodGet, "/pool/id/"+id, nil, &resp)
	return resp, err
}

func (c *Client) DeletePool(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/pool/"+id, nil, nil)
}

type AddressResult struct {
	Address string `json:"address"`
	Gateway string `json:"gateway"`
	LeaseID string `json:"leaseId"`
	VLAN    *int64 `json:"vlan,omitempty"`
}

type AssignCreateRequest struct {
	Region []string `json:"region"`
	Host   string   `json:"host"`
	Key    string   `json:"key"`
	Type   string   `json:"type"`
	Data   any      `json:"data,omitempty"`
}

type AssignCreateResponse struct {
	ID        string                   `json:"id"`
	Addresses map[string]AddressResult `json:"addresses"`
}

type AssignInfo struct {
	ID        string                   `json:"id"`
	CreatedAt string                   `json:"createdAt"`
	Key       string                   `json:"key"`
	Type      string                   `json:"type"`
	Data      any                      `json:"data"`
	Leases    map[string]AddressResult `json:"leases"`
}

func (c *Client) CreateAssign(ctx context.Context, req AssignCreateRequest) (AssignCreateResponse, error) {
	var resp AssignCreateResponse
	err := c.do(ctx, http.MethodPost, "/assigns", req, &resp)
	return resp, err
}

func (c *Client) GetAssign(ctx context.Context, id string) (AssignInfo, error) {
	var resp AssignInfo
	err := c.do(ctx, http.MethodGet, "/assign/"+id, nil, &resp)
	return resp, err
}

func (c *Client) DeleteAssign(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/assign/"+id, nil, nil)
}
