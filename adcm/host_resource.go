package adcm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	adcmClient "github.com/giggsoff/terraform-provider-adcm/client"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &hostResource{}
	_ resource.ResourceWithConfigure   = &hostResource{}
	_ resource.ResourceWithImportState = &hostResource{}
)

// NewHostResource is a helper function to simplify the provider implementation.
func NewHostResource() resource.Resource {
	return &hostResource{}
}

// hostResource is the resource implementation.
type hostResource struct {
	client *adcmClient.Client
}

// hostResourceModel maps order item data.
type hostResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	FQDN        types.String `tfsdk:"fqdn"`
	Description types.String `tfsdk:"description"`
	ProviderID  types.Int64  `tfsdk:"provider_id"`
	ClusterID   types.Int64  `tfsdk:"cluster_id"`
	Config      types.String `tfsdk:"config"`
	ConfigApply types.String `tfsdk:"config_apply"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

// Metadata returns the data source type name.
func (r *hostResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

// Schema defines the schema for the data source.
func (r *hostResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an order.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the order.",
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the order.",
				Computed:    true,
			},
			"fqdn": schema.StringAttribute{
				Description: "FQDN of host.",
				Required:    true,
			},
			"provider_id": schema.Int64Attribute{
				Description: "Provider ID of host.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "FQDN of host.",
				Optional:    true,
			},
			"cluster_id": schema.Int64Attribute{
				Description: "Cluster ID of host.",
				Optional:    true,
			},
			"config_apply": schema.StringAttribute{
				Description: "Config of host in JSON string to apply.",
				Optional:    true,
			},
			"config": schema.StringAttribute{
				Description: "Config of host in JSON string from ADCM.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *hostResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*adcmClient.Client)
}

// Create creates the resource and sets the initial Terraform state.
func (r *hostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan hostResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var host adcmClient.Host
	host.ClusterID = plan.ClusterID.ValueInt64()
	host.ProviderID = plan.ProviderID.ValueInt64()
	host.FQDN = plan.FQDN.ValueString()
	host.Description = plan.Description.ValueString()
	if plan.ConfigApply.ValueString() != "" {
		err := json.Unmarshal([]byte(plan.ConfigApply.ValueString()), &host.Config)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating host",
				"Could not create host, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// Create new host
	h, err := r.client.CreateHost(host)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating host",
			"Could not create host, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.Int64Value(h.ID)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	cfg, err := json.Marshal(h.Config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading ADCM host config",
			fmt.Sprintf("Could not read ADCM host config ID %d: %s", h.ID, err),
		)
		return
	}
	plan.Config = types.StringValue(string(cfg))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *hostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state hostResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed host value from ADCM
	h, err := r.client.GetHost(adcmClient.HostSearch{Identifier: adcmClient.Identifier{ID: state.ID.ValueInt64()}})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading ADCM host",
			fmt.Sprintf("Could not read ADCM host ID %d: %s", state.ID.ValueInt64(), err),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.Int64Value(h.ID)
	state.FQDN = types.StringValue(h.FQDN)
	state.Description = types.StringValue(h.Description)
	state.ProviderID = types.Int64Value(h.ProviderID)
	state.ClusterID = types.Int64Value(h.ClusterID)
	cfg, err := json.Marshal(h.Config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading ADCM host config",
			fmt.Sprintf("Could not read ADCM host config ID %d: %s", state.ID.ValueInt64(), err),
		)
		return
	}
	state.Config = types.StringValue(string(cfg))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *hostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	panic("implement me")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *hostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state hostResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing host
	err := r.client.DeleteHost(adcmClient.HostSearch{Identifier: adcmClient.Identifier{ID: state.ID.ValueInt64()}})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting ADCM host",
			"Could not delete host, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *hostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
