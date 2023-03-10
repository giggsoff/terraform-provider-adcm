package adcm

import (
	"context"
	adcmClient "github.com/giggsoff/terraform-provider-adcm/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &providerDataSource{}
	_ datasource.DataSourceWithConfigure = &providerDataSource{}
)

// NewProviderDataSource is a helper function to simplify the provider implementation.
func NewProviderDataSource() datasource.DataSource {
	return &providerDataSource{}
}

// providerDataSource is the data source implementation.
type providerDataSource struct {
	client *adcmClient.Client
}

// bundleModel maps bundle schema data.
type providerDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	BundleID    types.Int64  `tfsdk:"bundle_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	State       types.String `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (d *providerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider"
}

// Schema defines the schema for the data source.
func (d *providerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the provider.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the provider.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Product name of the provider.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Product description of the provider.",
				Optional:    true,
				Computed:    true,
			},
			"bundle_id": schema.Int64Attribute{
				Description: "Numeric identifier of the provider's bundle.",
				Optional:    true,
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "State of the provider's bundle.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *providerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*adcmClient.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *providerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state providerDataSourceModel

	var requestOptions providerDataSourceModel

	req.Config.Get(ctx, &requestOptions)

	opts := adcmClient.ProviderSearch{
		Name:        requestOptions.Name.ValueString(),
		Description: requestOptions.Description.ValueString(),
		BundleID:    requestOptions.BundleID.ValueInt64(),
		State:       requestOptions.State.ValueString(),
	}

	provider, err := d.client.GetProvider(opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read ADCM Provider",
			err.Error(),
		)
		return
	}

	state.ID = types.Int64Value(int64(provider.ID))
	state.Name = types.StringValue(provider.Name)
	state.Description = types.StringValue(provider.Description)
	state.BundleID = types.Int64Value(provider.BundleID)
	state.State = types.StringValue(provider.State)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
