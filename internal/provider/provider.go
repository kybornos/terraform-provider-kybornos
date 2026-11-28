package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kybornos/terraform-provider-kybornos/internal/client"
	"os"
)

var _ provider.Provider = (*KybornosProvider)(nil)

type KybornosProvider struct{ version string }
type providerModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	APIKey   types.String `tfsdk:"api_key"`
}

func New() provider.Provider { return &KybornosProvider{version: "dev"} }
func (p *KybornosProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kybornos"
	resp.Version = p.version
}
func (p *KybornosProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terraform provider for Kybornos VDC infrastructure.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Kybornos Gateway API base URL. " +
					"Can also be set using the `KYBORNOS_ENDPOINT` environment variable.",
			},
			"api_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				MarkdownDescription: "Kybornos Gateway API key. " +
					"Can also be set using the `KYBORNOS_API_KEY` environment variable.",
			},
		},
	}
}
func (p *KybornosProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var m providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	endpoint := os.Getenv("KYBORNOS_ENDPOINT")
	if !m.Endpoint.IsNull() && !m.Endpoint.IsUnknown() {
		endpoint = m.Endpoint.ValueString()
	}
	if endpoint == "" {
		endpoint = "https://10.0.0.1:8443"
	}
	key := os.Getenv("KYBORNOS_API_KEY")
	if !m.APIKey.IsNull() && !m.APIKey.IsUnknown() {
		key = m.APIKey.ValueString()
	}
	if m.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("endpoint"), "Unknown endpoint", "Endpoint must be known during provider configuration.")
		return
	}
	if m.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("api_key"), "Unknown API key", "API key must be known during provider configuration.")
		return
	}
	c, err := client.New(endpoint, key)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Kybornos configuration", err.Error())
		return
	}
	resp.ResourceData = c
	resp.DataSourceData = c
}
func (p *KybornosProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{NewVMResource}
}
func (p *KybornosProvider) DataSources(_ context.Context) []func() datasource.DataSource { return nil }
