// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

var _ function.Function = &IPv4Long2IPFunction{}

func NewIPv4Long2IPFunction() function.Function {
	return &IPv4Long2IPFunction{}
}

type IPv4Long2IPFunction struct{}

func (f *IPv4Long2IPFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "ipv4_long2ip"
}

func (f *IPv4Long2IPFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Convert IPv4 number to dotted string",
		Description: "Converts an IPv4 address from its 32-bit integer representation to dotted-decimal string form.",
		Parameters: []function.Parameter{
			function.Int64Parameter{
				Name:        "ipv4_long",
				Description: "IPv4 address as 32-bit integer in range [0, 4294967295]",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *IPv4Long2IPFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var ipv4Long int64
	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &ipv4Long))
	if resp.Error != nil {
		return
	}

	ip, err := ipv4LongToIP(ipv4Long)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, err.Error()))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, ip))
}

var _ function.Function = &IPv4IP2LongFunction{}

func NewIPv4IP2LongFunction() function.Function {
	return &IPv4IP2LongFunction{}
}

type IPv4IP2LongFunction struct{}

func (f *IPv4IP2LongFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "ipv4_ip2long"
}

func (f *IPv4IP2LongFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Convert IPv4 dotted string to number",
		Description: "Converts an IPv4 address from dotted-decimal string form to its 32-bit integer representation.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "ipv4_ip",
				Description: "IPv4 address in dotted-decimal form, e.g. 192.168.1.1",
			},
		},
		Return: function.Int64Return{},
	}
}

func (f *IPv4IP2LongFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var ipv4IP string
	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &ipv4IP))
	if resp.Error != nil {
		return
	}

	result, err := ipv4IPToLong(ipv4IP)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, err.Error()))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, &result))
}

func ipv4LongToIP(value int64) (string, error) {
	if value < 0 || value > math.MaxUint32 {
		return "", fmt.Errorf("ipv4_long must be between 0 and %d", uint64(math.MaxUint32))
	}

	var b [4]byte
	binary.BigEndian.PutUint32(b[:], uint32(value))
	return net.IPv4(b[0], b[1], b[2], b[3]).String(), nil
}

func ipv4IPToLong(value string) (int64, error) {
	ip := net.ParseIP(value)
	if ip == nil {
		return 0, fmt.Errorf("ipv4_ip must be a valid IPv4 address")
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return 0, fmt.Errorf("ipv4_ip must be a valid IPv4 address")
	}
	return int64(binary.BigEndian.Uint32(ip4)), nil
}
