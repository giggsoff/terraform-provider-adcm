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
	_ resource.Resource                = &clusterResource{}
	_ resource.ResourceWithConfigure   = &clusterResource{}
	_ resource.ResourceWithImportState = &clusterResource{}
)

// NewClusterResource is a helper function to simplify the provider implementation.
func NewClusterResource() resource.Resource {
	return &clusterResource{}
}

// clusterResource is the resource implementation.
type clusterResource struct {
	client *adcmClient.Client
}

// clusterResourceModel maps order item data.
type clusterResourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	BundleID       types.Int64  `tfsdk:"bundle_id"`
	ClusterConfig  types.String `tfsdk:"cluster_config"`
	ServicesConfig types.String `tfsdk:"services_config"`
	HCMap          types.String `tfsdk:"hc_map"`
}

// Metadata returns the data source type name.
func (r *clusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

// Schema defines the schema for the data source.
func (r *clusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "name of cluster.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "FQDN of cluster.",
				Required:    true,
			},
			"bundle_id": schema.Int64Attribute{
				Description: "Cluster ID of cluster.",
				Required:    true,
			},
			"cluster_config": schema.StringAttribute{
				Description: "Config of cluster in JSON string to apply.",
				Optional:    true,
			},
			"services_config": schema.StringAttribute{
				Description: "Config of services in JSON string to apply.",
				Optional:    true,
			},
			"hc_map": schema.StringAttribute{
				Description: "Config of host-component mapping in JSON string to apply.",
				Optional:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *clusterResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*adcmClient.Client)
}

// Create creates the resource and sets the initial Terraform state.
func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan clusterResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var cluster adcmClient.Cluster
	cluster.BundleID = plan.BundleID.ValueInt64()
	cluster.Name = plan.Name.ValueString()
	cluster.Description = plan.Description.ValueString()
	if plan.ClusterConfig.ValueString() != "" {
		err := json.Unmarshal([]byte(plan.ClusterConfig.ValueString()), &cluster.ClusterConfig.Config)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating cluster",
				"Could not unmarshal cluster config of cluster, unexpected error: "+err.Error(),
			)
			return
		}
	}
	if plan.ServicesConfig.ValueString() != "" {
		err := json.Unmarshal([]byte(plan.ServicesConfig.ValueString()), &cluster.ServicesConfig.Config)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating cluster",
				"Could not unmarshal services config of cluster, unexpected error: "+err.Error(),
			)
			return
		}
	}
	if plan.HCMap.ValueString() != "" {
		err := json.Unmarshal([]byte(plan.HCMap.ValueString()), &cluster.HCMap)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating cluster",
				"Could not unmarshal hc map of cluster, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// Create new cluster
	h, err := r.client.CreateCluster(cluster)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating cluster",
			"Could not create cluster, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.Int64Value(h.ID)
	plan.BundleID = types.Int64Value(h.BundleID)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state clusterResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed cluster value from ADCM
	h, err := r.client.GetCluster(adcmClient.ClusterSearch{Identifier: adcmClient.Identifier{ID: state.ID.ValueInt64()}})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading ADCM cluster",
			fmt.Sprintf("Could not read ADCM cluster ID %d: %s", state.ID.ValueInt64(), err),
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
func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error Update ADCM cluster",
		fmt.Sprintf("%+v", req),
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state clusterResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing cluster
	err := r.client.DeleteCluster(adcmClient.ClusterSearch{Identifier: adcmClient.Identifier{ID: state.ID.ValueInt64()}})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting ADCM cluster",
			"Could not delete cluster, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *clusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
