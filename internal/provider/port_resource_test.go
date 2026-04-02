// Copyright WANIX Inc. 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPortResourceAndDataSource(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/port":
			if got := r.Header.Get("X-Portd-Host"); got != "node-1" {
				http.Error(w, "unexpected X-Portd-Host", http.StatusBadRequest)
				return
			}

			var req struct {
				AssignID   string `json:"assignId"`
				TargetPort int64  `json:"targetPort"`
				Service    string `json:"service"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if req.AssignID != "assign-1" || req.TargetPort != 22 || req.Service != "ssh" {
				http.Error(w, "unexpected request payload", http.StatusBadRequest)
				return
			}

			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":   "port-1",
				"port": 32022,
			})
		case r.Method == http.MethodGet && r.URL.Path == "/port/id/port-1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":         "port-1",
				"assignId":   "assign-1",
				"host":       "node-1",
				"port":       32022,
				"targetPort": 22,
				"service":    "ssh",
				"createdAt":  "2024-01-01T00:00:00Z",
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/port/id/port-1":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccPortConfig(server.URL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zeus_port.test", "assign_id", "assign-1"),
					resource.TestCheckResourceAttr("zeus_port.test", "scope_host", "node-1"),
					resource.TestCheckResourceAttr("zeus_port.test", "host", "node-1"),
					resource.TestCheckResourceAttr("zeus_port.test", "port", "32022"),
					resource.TestCheckResourceAttr("zeus_port.test", "target_port", "22"),
					resource.TestCheckResourceAttr("zeus_port.test", "service", "ssh"),
					resource.TestCheckResourceAttr("data.zeus_port.by_id", "assign_id", "assign-1"),
					resource.TestCheckResourceAttr("data.zeus_port.by_id", "host", "node-1"),
					resource.TestCheckResourceAttr("data.zeus_port.by_id", "port", "32022"),
				),
			},
			{
				ResourceName:      "zeus_port.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"scope_host",
				},
			},
		},
	})
}

func TestAccPortResource_DefaultScopeImportStable(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/port":
			if got := r.Header.Get("X-Portd-Host"); got != "" {
				http.Error(w, "unexpected X-Portd-Host", http.StatusBadRequest)
				return
			}

			var req struct {
				AssignID   string `json:"assignId"`
				TargetPort int64  `json:"targetPort"`
				Service    string `json:"service"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":   "port-default",
				"port": 30080,
			})
		case r.Method == http.MethodGet && r.URL.Path == "/port/id/port-default":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":         "port-default",
				"assignId":   "assign-default",
				"host":       "Default",
				"port":       30080,
				"targetPort": 80,
				"service":    "http",
				"createdAt":  "2024-01-01T00:00:00Z",
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/port/id/port-default":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccPortDefaultScopeConfig(server.URL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("zeus_port.default_scope", "scope_host"),
					resource.TestCheckResourceAttr("zeus_port.default_scope", "host", "Default"),
					resource.TestCheckResourceAttr("zeus_port.default_scope", "port", "30080"),
				),
			},
			{
				ResourceName:      "zeus_port.default_scope",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPortResource_EmptyScopeHostOmitsHeader(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/port":
			if got := r.Header.Get("X-Portd-Host"); got != "" {
				http.Error(w, "unexpected X-Portd-Host", http.StatusBadRequest)
				return
			}

			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":   "port-empty-scope",
				"port": 30081,
			})
		case r.Method == http.MethodGet && r.URL.Path == "/port/id/port-empty-scope":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":         "port-empty-scope",
				"assignId":   "assign-empty",
				"host":       "Default",
				"port":       30081,
				"targetPort": 81,
				"service":    "http-alt",
				"createdAt":  "2024-01-01T00:00:00Z",
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/port/id/port-empty-scope":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories(),
		Steps: []resource.TestStep{{
			Config: testAccPortEmptyScopeConfig(server.URL),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("zeus_port.empty_scope", "scope_host", ""),
				resource.TestCheckResourceAttr("zeus_port.empty_scope", "host", "Default"),
				resource.TestCheckResourceAttr("zeus_port.empty_scope", "port", "30081"),
			),
		}},
	})
}

func TestAccPortResource_ReadNotFoundPlansRecreate(t *testing.T) {
	missing := false
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/port":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "port-missing", "port": 31022})
		case r.Method == http.MethodGet && r.URL.Path == "/port/id/port-missing":
			if missing {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":         "port-missing",
				"assignId":   "assign-1",
				"host":       "node-1",
				"port":       31022,
				"targetPort": 22,
				"service":    "ssh",
				"createdAt":  "2024-01-01T00:00:00Z",
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/port/id/port-missing":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccPortResourceOnlyConfig(server.URL),
			},
			{
				PreConfig: func() {
					missing = true
				},
				Config:             testAccPortResourceOnlyConfig(server.URL),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccPortResource_DeleteNotFound(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/port":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "port-delete-404", "port": 32022})
		case r.Method == http.MethodGet && r.URL.Path == "/port/id/port-delete-404":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":         "port-delete-404",
				"assignId":   "assign-1",
				"host":       "node-1",
				"port":       32022,
				"targetPort": 22,
				"service":    "ssh",
				"createdAt":  "2024-01-01T00:00:00Z",
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/port/id/port-delete-404":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
		default:
			http.NotFound(w, r)
		}
	}))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories(),
		Steps: []resource.TestStep{{
			Config: testAccPortConfig(server.URL),
		}},
	})
}

func TestAccPortDataSource_NotFound(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}

		if r.Method == http.MethodGet && r.URL.Path == "/port/id/missing-port" {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
			return
		}

		http.NotFound(w, r)
	}))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { TestAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProviderFactories(),
		Steps: []resource.TestStep{{
			Config:      testAccPortDataSourceConfig(server.URL, "missing-port"),
			ExpectError: regexp.MustCompile(`(?s)Read port failed.*status 404: not found`),
		}},
	})
}

func testAccPortConfig(endpoint string) string {
	return `
provider "zeus" {
  endpoint = "` + endpoint + `"
  token    = "token"
}

resource "zeus_port" "test" {
  assign_id   = "assign-1"
  scope_host  = "node-1"
  target_port = 22
  service     = "ssh"
}

data "zeus_port" "by_id" {
  id = zeus_port.test.id
}
`
}

func testAccPortResourceOnlyConfig(endpoint string) string {
	return `
provider "zeus" {
  endpoint = "` + endpoint + `"
  token    = "token"
}

resource "zeus_port" "test" {
  assign_id   = "assign-1"
  scope_host  = "node-1"
  target_port = 22
  service     = "ssh"
}
`
}

func testAccPortDefaultScopeConfig(endpoint string) string {
	return `
provider "zeus" {
  endpoint = "` + endpoint + `"
  token    = "token"
}

resource "zeus_port" "default_scope" {
  assign_id   = "assign-default"
  target_port = 80
  service     = "http"
}
`
}

func testAccPortEmptyScopeConfig(endpoint string) string {
	return `
provider "zeus" {
  endpoint = "` + endpoint + `"
  token    = "token"
}

resource "zeus_port" "empty_scope" {
  assign_id   = "assign-empty"
  scope_host  = ""
  target_port = 81
  service     = "http-alt"
}
`
}

func testAccPortDataSourceConfig(endpoint, id string) string {
	return `
provider "zeus" {
  endpoint = "` + endpoint + `"
  token    = "token"
}

data "zeus_port" "missing" {
  id = "` + id + `"
}
`
}
