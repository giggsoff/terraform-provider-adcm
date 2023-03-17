package adcm

import (
	"context"
	"os"

	adcmClient "github.com/giggsoff/terraform-provider-adcm/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &adcmProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &adcmProvider{}
}

// adcmProvider is the provider implementation.
type adcmProvider struct{}

type adcmProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (a adcmProvider) Metadata(_ context.Context, _ provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "adcm"
}

func (a adcmProvider) Schema(_ context.Context, _ provider.SchemaRequest, response *provider.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Interact with ADCM.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "URI for ADCM API. May also be provided via ADCM_HOST environment variable.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username for ADCM API. May also be provided via ADCM_USERNAME environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for ADCM API. May also be provided via ADCM_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (a adcmProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring ADCM client")

	// Retrieve provider data from configuration
	var config adcmProviderModel
	diags := request.Config.Get(ctx, &config)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		response.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown ADCM API Host",
			"The provider cannot create the ADCM API client as there is an unknown configuration value for the ADCM API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ADCM_HOST environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		response.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown ADCM API Username",
			"The provider cannot create the ADCM API client as there is an unknown configuration value for the ADCM API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ADCM_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		response.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown ADCM API Password",
			"The provider cannot create the ADCM API client as there is an unknown configuration value for the ADCM API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ADCM_PASSWORD environment variable.",
		)
	}

	if response.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("ADCM_HOST")
	username := os.Getenv("ADCM_USERNAME")
	password := os.Getenv("ADCM_PASSWORD")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		response.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing ADCM API Host",
			"The provider cannot create the ADCM API client as there is a missing or empty value for the ADCM API host. "+
				"Set the host value in the configuration or use the ADCM_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		response.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing ADCM API Username",
			"The provider cannot create the ADCM API client as there is a missing or empty value for the ADCM API username. "+
				"Set the username value in the configuration or use the ADCM_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		response.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing ADCM API Password",
			"The provider cannot create the ADCM API client as there is a missing or empty value for the ADCM API password. "+
				"Set the password value in the configuration or use the ADCM_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if response.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "adcm_host", host)
	ctx = tflog.SetField(ctx, "adcm_username", username)
	ctx = tflog.SetField(ctx, "adcm_password", password)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "adcm_password")

	tflog.Debug(ctx, "Creating ADCM client")

	// Create a new ADCM client using the configuration values
	client, err := adcmClient.NewClient(&host, &username, &password)
	if err != nil {
		response.Diagnostics.AddError(
			"Unable to Create ADCM API Client",
			"An unexpected error occurred when creating the ADCM API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"ADCM Client Error: "+err.Error(),
		)
		return
	}

	// Make the ADCM client available during DataSource and Resource
	// type Configure methods.
	response.DataSourceData = client
	response.ResourceData = client

	tflog.Info(ctx, "Configured ADCM client", map[string]any{"success": true})
}

func (a adcmProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewBundleDataSource,
		NewProviderDataSource,
	}
}

func (a adcmProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewHostResource,
		NewClusterResource,
		NewBundleResource,
	}
}
