// Copyright (c) WANIX Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestAccProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"zeus": providerserver.NewProtocol6WithError(New("test")()),
	}
}

func TestAccPreCheck(t *testing.T) {
	// Placeholder for required environment assertions.
}

func newTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("skip: cannot open test listener: %v", err)
		return nil
	}
	server := httptest.NewUnstartedServer(handler)
	server.Listener = ln
	server.Start()
	t.Cleanup(server.Close)
	return server
}
