// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIPv4Long2IPFunctionRun(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  function.RunRequest
		expected function.RunResponse
	}{
		"value-valid": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.Int64Value(3232235777)}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("192.168.1.1")),
			},
		},
		"value-invalid-negative": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.Int64Value(-1)}),
			},
			expected: function.RunResponse{
				Error:  function.NewArgumentFuncError(0, "ipv4_long must be between 0 and 4294967295"),
				Result: function.NewResultData(types.StringUnknown()),
			},
		},
		"value-invalid-overflow": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.Int64Value(4294967296)}),
			},
			expected: function.RunResponse{
				Error:  function.NewArgumentFuncError(0, "ipv4_long must be between 0 and 4294967295"),
				Result: function.NewResultData(types.StringUnknown()),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := function.RunResponse{
				Result: function.NewResultData(types.StringUnknown()),
			}

			(&IPv4Long2IPFunction{}).Run(context.Background(), testCase.request, &got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestIPv4IP2LongFunctionRun(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  function.RunRequest
		expected function.RunResponse
	}{
		"value-valid": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("192.168.1.1")}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.Int64Value(3232235777)),
			},
		},
		"value-invalid": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("not-an-ip")}),
			},
			expected: function.RunResponse{
				Error:  function.NewArgumentFuncError(0, "ipv4_ip must be a valid IPv4 address"),
				Result: function.NewResultData(types.Int64Unknown()),
			},
		},
		"value-invalid-ipv6": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("::1")}),
			},
			expected: function.RunResponse{
				Error:  function.NewArgumentFuncError(0, "ipv4_ip must be a valid IPv4 address"),
				Result: function.NewResultData(types.Int64Unknown()),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := function.RunResponse{
				Result: function.NewResultData(types.Int64Unknown()),
			}

			(&IPv4IP2LongFunction{}).Run(context.Background(), testCase.request, &got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
