package adcm

import (
	"context"
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
	_ resource.Resource                = &bundleResource{}
	_ resource.ResourceWithConfigure   = &bundleResource{}
	_ resource.ResourceWithImportState = &bundleResource{}
)

// NewBundleResource is a helper function to simplify the provider implementation.
func NewBundleResource() resource.Resource {
	return &bundleResource{}
}

// bundleResource is the resource implementation.
type bundleResource struct {
	client *adcmClient.Client
}

// bundleModel maps order item data.
type bundleModel struct {
	ID      types.Int64  `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
	Edition types.String `tfsdk:"edition"`
	URL     types.String `tfsdk:"url"`
}

// Metadata returns the data source type name.
func (r *bundleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bundle"
}

// Schema defines the schema for the data source.
func (r *bundleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an bundle.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Numeric identifier of the bundle.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "name of bundle.",
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "version of bundle.",
				Computed:    true,
			},
			"edition": schema.StringAttribute{
				Description: "edition of bundle.",
				Computed:    true,
			},
			"url": schema.StringAttribute{
				Description: "URL of bundle.",
				Required:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *bundleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*adcmClient.Client)
}

// Create creates the resource and sets the initial Terraform state.
func (r *bundleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan bundleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bundle, err := r.client.UploadBundle(plan.URL.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating bundle",
			"Could not upload bundle, unexpected error: "+err.Error(),
		)
		return
	}
	plan.ID = types.Int64Value(bundle.ID)
	plan.Name = types.StringValue(bundle.Name)
	plan.Version = types.StringValue(bundle.Version)
	plan.Edition = types.StringValue(bundle.Edition)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *bundleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state bundleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bundle, err := r.client.GetBundle(adcmClient.BundleSearch{Identifier: adcmClient.Identifier{ID: state.ID.ValueInt64()}})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read ADCM Bundle",
			err.Error(),
		)
		return
	}

	state.ID = types.Int64Value(bundle.ID)
	state.Name = types.StringValue(bundle.Name)
	state.Edition = types.StringValue(bundle.Edition)
	state.Version = types.StringValue(bundle.Version)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *bundleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error Update ADCM bundle",
		fmt.Sprintf("%+v", req),
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *bundleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state bundleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing bundle
	err := r.client.DeleteBundle(adcmClient.BundleSearch{Identifier: adcmClient.Identifier{ID: state.ID.ValueInt64()}})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting ADCM bundle",
			"Could not delete bundle, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *bundleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
