// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/5aaee9/terraform-provider-zeus/internal/zeusapi"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPoolResourceAndDataSource(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/pools":
			var req zeusapi.CreatePoolRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(zeusapi.CreatePoolResponse{ID: "pool-1"})
		case r.Method == http.MethodGet && r.URL.Path == "/pool/id/pool-1":
			_ = json.NewEncoder(w).Encode(zeusapi.PoolDetail{
				ID:           "pool-1",
				Region:       "us-east-1",
				FriendlyName: "primary",
				Begin:        "10.0.0.1",
				End:          "10.0.0.9",
				Gateway:      "10.0.0.254",
				State:        []int64{0, 1, 0},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/pool/pool-1":
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
				Config: testAccPoolConfig(server.URL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zeus_pool.test", "region", "us-east-1"),
					resource.TestCheckResourceAttr("zeus_pool.test", "friendly_name", "primary"),
					resource.TestCheckResourceAttr("zeus_pool.test", "state.#", "3"),
					resource.TestCheckResourceAttr("data.zeus_pool.by_id", "gateway_ip", "10.0.0.254"),
					resource.TestCheckResourceAttr("data.zeus_pool.by_id", "size", "3"),
				),
			},
			{
				ResourceName:      "zeus_pool.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPoolConfig(endpoint string) string {
	return `
provider "zeus" {
  endpoint = "` + endpoint + `"
  token    = "token"
}

resource "zeus_pool" "test" {
  start   = 1
  gateway = 2
  size    = 3
  region  = "us-east-1"
}

data "zeus_pool" "by_id" {
  id = zeus_pool.test.id
}
`
}
