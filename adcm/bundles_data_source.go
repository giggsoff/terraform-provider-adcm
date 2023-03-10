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
	_ datasource.DataSource              = &bundleDataSource{}
	_ datasource.DataSourceWithConfigure = &bundleDataSource{}
)

// NewBundleDataSource is a helper function to simplify the provider implementation.
func NewBundleDataSource() datasource.DataSource {
	return &bundleDataSource{}
}

// bundleDataSource is the data source implementation.
type bundleDataSource struct {
	client *adcmClient.Client
}

// bundleModel maps bundle schema data.
type bundleDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	Edition     types.String `tfsdk:"edition"`
	License     types.String `tfsdk:"license"`
	Version     types.String `tfsdk:"version"`
}

// Metadata returns the data source type name.
func (d *bundleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bundle"
}

// Schema defines the schema for the data source.
func (d *bundleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the bundle.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the bundle.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Product name of the bundle.",
				Optional:    true,
				Computed:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "Product display name of the bundle.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Product description of the bundle.",
				Optional:    true,
				Computed:    true,
			},
			"edition": schema.StringAttribute{
				Description: "Product edition of the bundle.",
				Optional:    true,
				Computed:    true,
			},
			"license": schema.StringAttribute{
				Description: "Product license acceptance state of the bundle.",
				Optional:    true,
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "Product version of the bundle.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *bundleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*adcmClient.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *bundleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state bundleDataSourceModel

	var requestOptions bundleDataSourceModel

	req.Config.Get(ctx, &requestOptions)

	opts := adcmClient.BundleSearch{
		Name:        requestOptions.Name.ValueString(),
		DisplayName: requestOptions.DisplayName.ValueString(),
		Description: requestOptions.Description.ValueString(),
		Edition:     requestOptions.Edition.ValueString(),
		License:     requestOptions.License.ValueString(),
		Version:     requestOptions.Version.ValueString(),
	}
	bundle, err := d.client.GetBundle(opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read ADCM Bundle",
			err.Error(),
		)
		return
	}

	state.ID = types.Int64Value(int64(bundle.ID))
	state.Name = types.StringValue(bundle.Name)
	state.DisplayName = types.StringValue(bundle.DisplayName)
	state.Description = types.StringValue(bundle.Description)
	state.Edition = types.StringValue(bundle.Edition)
	state.License = types.StringValue(bundle.License)
	state.Version = types.StringValue(bundle.Version)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
