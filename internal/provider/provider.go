// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/5aaee9/terraform-provider-zeus/internal/zeusapi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ZeusProvider satisfies various provider interfaces.
var _ provider.Provider = &ZeusProvider{}
var _ provider.ProviderWithFunctions = &ZeusProvider{}

// ZeusProvider defines the provider implementation.
type ZeusProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ZeusProviderModel describes the provider data model.
type ZeusProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

func (p *ZeusProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "zeus"
	resp.Version = p.version
}

func (p *ZeusProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Zeus API endpoint, e.g. http://host:port",
				Required:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Bearer token for Zeus API",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *ZeusProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ZeusProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := zeusapi.NewClient(data.Endpoint.ValueString(), data.Token.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Invalid client configuration", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ZeusProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPoolResource,
		NewAssignResource,
	}
}

func (p *ZeusProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPoolDataSource,
		NewAssignDataSource,
	}
}

func (p *ZeusProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ZeusProvider{
			version: version,
		}
	}
}
