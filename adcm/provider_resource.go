package adcm

import (
	"context"
	"encoding/json"
	"fmt"

	adcmClient "github.com/giggsoff/terraform-provider-adcm/client"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &providerResource{}
	_ resource.ResourceWithConfigure   = &providerResource{}
	_ resource.ResourceWithImportState = &providerResource{}
)

// NewProviderResource is a helper function to simplify the provider implementation.
func NewProviderResource() resource.Resource {
	return &providerResource{}
}

// providerResource is the resource implementation.
type providerResource struct {
	client *adcmClient.Client
}

// providerResourceModel maps order item data.
type providerResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	BundleID    types.Int64  `tfsdk:"bundle_id"`
	Config      types.String `tfsdk:"config"`
}

// Metadata returns the data source type name.
func (r *providerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider"
}

// Schema defines the schema for the data source.
func (r *providerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an provider.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the provider.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "name of provider.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "FQDN of provider.",
				Required:    true,
			},
			"bundle_id": schema.Int64Attribute{
				Description: "Bundle ID of provider.",
				Required:    true,
			},
			"config": schema.StringAttribute{
				Description: "Config of provider in JSON string to apply.",
				Optional:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *providerResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*adcmClient.Client)
}

// Create creates the resource and sets the initial Terraform state.
func (r *providerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan providerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var provider adcmClient.Provider
	provider.BundleID = plan.BundleID.ValueInt64()
	provider.Name = plan.Name.ValueString()
	provider.Description = plan.Description.ValueString()
	if plan.Config.ValueString() != "" {
		err := json.Unmarshal([]byte(plan.Config.ValueString()), &provider.ProviderConfig.Config)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating provider",
				"Could not unmarshal provider config of provider, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// Create new provider
	p, err := r.client.CreateProvider(provider)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating provider",
			"Could not create provider, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.Int64Value(p.ID)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *providerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state providerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed provider value from ADCM
	h, err := r.client.GetProvider(adcmClient.ProviderSearch{Identifier: adcmClient.Identifier{ID: state.ID.ValueInt64()}})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading ADCM provider",
			fmt.Sprintf("Could not read ADCM provider ID %d: %s", state.ID.ValueInt64(), err),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.Int64Value(h.ID)
	state.Name = types.StringValue(h.Name)
	if h.Description != state.Description.ValueString() {
		state.Description = types.StringValue(h.Description)
	}
	state.BundleID = types.Int64Value(h.BundleID)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *providerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error Update ADCM provider",
		fmt.Sprintf("%+v", req),
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *providerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state providerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing provider
	err := r.client.DeleteProvider(adcmClient.ProviderSearch{Identifier: adcmClient.Identifier{ID: state.ID.ValueInt64()}})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting ADCM provider",
			"Could not delete provider, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *providerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
