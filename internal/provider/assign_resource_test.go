// Copyright (c) WANIX Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/5aaee9/terraform-provider-zeus/internal/zeusapi"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAssignResourceAndDataSource(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/assigns":
			var req zeusapi.AssignCreateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(zeusapi.AssignCreateResponse{
				ID: "assign-1",
				Addresses: map[string]zeusapi.AddressResult{
					"us-east-1": {
						Address: "10.0.0.5",
						Gateway: "10.0.0.254",
						LeaseID: "lease-1",
					},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/assign/assign-1":
			_ = json.NewEncoder(w).Encode(zeusapi.AssignInfo{
				ID:        "assign-1",
				CreatedAt: "2024-01-01T00:00:00Z",
				Key:       "vm-1",
				Type:      "vm",
				Data: map[string]interface{}{
					"tag": "blue",
				},
				Leases: map[string]zeusapi.AddressResult{
					"us-east-1": {
						Address: "10.0.0.5",
						Gateway: "10.0.0.254",
						LeaseID: "lease-1",
					},
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/assign/assign-1":
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
				Config: testAccAssignConfig(server.URL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zeus_assign.test", "key", "vm-1"),
					resource.TestCheckResourceAttr("zeus_assign.test", "leases.us-east-1.address", "10.0.0.5"),
					resource.TestCheckResourceAttr("data.zeus_assign.by_id", "type", "vm"),
					resource.TestCheckResourceAttr("data.zeus_assign.by_id", "leases.us-east-1.gateway", "10.0.0.254"),
				),
			},
			{
				ResourceName:      "zeus_assign.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAssignConfig(endpoint string) string {
	return `
provider "zeus" {
  endpoint = "` + endpoint + `"
  token    = "token"
}

resource "zeus_assign" "test" {
  region = ["us-east-1"]
  host   = "host-1"
  key    = "vm-1"
  type   = "vm"
  data   = { tag = "blue" }
}

data "zeus_assign" "by_id" {
  id = zeus_assign.test.id
}
`
}
