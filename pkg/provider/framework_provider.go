package provider

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/resources"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MaterializeProvider struct {
	// Define provider configuration and internal client here
	version string
	client  *utils.ProviderMeta
}

type providerData struct {
	Endpoint      types.String `tfsdk:"endpoint"`
	CloudEndpoint types.String `tfsdk:"cloud_endpoint"`
	BaseEndpoint  types.String `tfsdk:"base_endpoint"`
	DefaultRegion types.String `tfsdk:"default_region"`
	Password      types.String `tfsdk:"password"`
	Database      types.String `tfsdk:"database"`
	SslMode       types.String `tfsdk:"sslmode"`
}

// Ensure MaterializeProvider satisfies various provider interfaces.
var _ provider.Provider = new(MaterializeProvider)

func New(version string) provider.Provider {
	return &MaterializeProvider{
		version: version,
	}
}

func (p *MaterializeProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "materialize"
	resp.Version = p.version
}

func (p *MaterializeProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"password": schema.StringAttribute{
				Description: "Materialize host. Can also come from the `MZ_PASSWORD` environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"database": schema.StringAttribute{
				Description: "The Materialize database. Can also come from the `MZ_DATABASE` environment variable. Defaults to `materialize`.",
				Optional:    true,
			},
			"sslmode": schema.StringAttribute{
				Description: "For testing purposes, the SSL mode to use.",
				Optional:    true,
			},
			"endpoint": schema.StringAttribute{
				Description: "The endpoint for the Materialize API.",
				Optional:    true,
			},
			"cloud_endpoint": schema.StringAttribute{
				Description: "The endpoint for the Materialize Cloud API.",
				Optional:    true,
			},
			"base_endpoint": schema.StringAttribute{
				Description: "The base endpoint for Materialize.",
				Optional:    true,
			},
			"default_region": schema.StringAttribute{
				Description: "The default region if not specified in the resource",
				Optional:    true,
			},
		},
	}
}

func (p *MaterializeProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewClusterResource,
	}
}

func (p *MaterializeProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// Configure implements the logic from your providerConfigure function adapted for the Plugin Framework
func (p *MaterializeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerData

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extracting values from providerData or falling back to environment variables
	password := config.Password.ValueString()
	if password == "" {
		password = os.Getenv("MZ_PASSWORD")
	}

	database := config.Database.ValueString()
	if database == "" {
		database = os.Getenv("MZ_DATABASE")
		if database == "" {
			database = "materialize"
		}
	}

	sslMode := config.SslMode.ValueString()
	if sslMode == "" {
		sslMode = os.Getenv("MZ_SSLMODE")
		if sslMode == "" {
			sslMode = "require"
		}
	}

	endpoint := config.Endpoint.ValueString()
	if endpoint == "" {
		endpoint = os.Getenv("MZ_ENDPOINT")
		if endpoint == "" {
			endpoint = "https://admin.cloud.materialize.com"
		}
	}

	cloudEndpoint := config.CloudEndpoint.ValueString()
	if cloudEndpoint == "" {
		cloudEndpoint = os.Getenv("MZ_CLOUD_ENDPOINT")
		if cloudEndpoint == "" {
			cloudEndpoint = "https://api.cloud.materialize.com"
		}
	}

	baseEndpoint := config.BaseEndpoint.ValueString()
	if baseEndpoint == "" {
		baseEndpoint = os.Getenv("MZ_BASE_ENDPOINT")
		if baseEndpoint == "" {
			baseEndpoint = "https://cloud.materialize.com"
		}
	}

	defaultRegion := config.DefaultRegion.ValueString()
	if defaultRegion == "" {
		defaultRegion = os.Getenv("MZ_DEFAULT_REGION")
		if defaultRegion == "" {
			defaultRegion = "aws/us-east-1"
		}
	}

	applicationName := fmt.Sprintf("terraform-provider-materialize v%s", p.version)

	err := utils.SetDefaultRegion(defaultRegion)
	if err != nil {
		resp.Diagnostics.AddError("Failed to set default region", err.Error())
		return
	}

	// Initialize the Frontegg client
	fronteggClient, err := clients.NewFronteggClient(ctx, password, endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Frontegg client", err.Error())
		return
	}

	// Initialize the Cloud API client using the Frontegg client and endpoint
	cloudAPIClient := clients.NewCloudAPIClient(fronteggClient, cloudEndpoint, baseEndpoint)
	regionsEnabled := make(map[clients.Region]bool)

	// Get the list of cloud providers
	providers, err := cloudAPIClient.ListCloudProviders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list cloud providers", err.Error())
		return
	}

	// Store the DB clients for all regions
	dbClients := make(map[clients.Region]*clients.DBClient)
	for _, provider := range providers {
		regionDetails, err := cloudAPIClient.GetRegionDetails(ctx, provider)
		log.Printf("[DEBUG] Region details for provider %s: %v\n", provider.ID, regionDetails)

		if err != nil {
			log.Printf("[ERROR] Error getting region details for provider %s: %v\n", provider.ID, err)
			continue
		}

		if regionDetails == nil || regionDetails.RegionInfo == nil {
			continue
		}

		regionsEnabled[clients.Region(provider.ID)] = regionDetails.RegionInfo != nil && regionDetails.RegionInfo.Resolvable

		// Get the database connection details for the region
		host, port, err := clients.SplitHostPort(regionDetails.RegionInfo.SqlAddress)
		if err != nil {
			log.Printf("[ERROR] Error splitting host and port for region %s: %v\n", provider.ID, err)
			continue
		}

		user := fronteggClient.Email

		// Instantiate a new DB client for the region
		dbClient, diags := clients.NewDBClient(host, user, password, port, database, applicationName, p.version, sslMode)
		if diags.HasError() {
			log.Printf("[ERROR] Error initializing DB client for region %s: %v\n", provider.ID, diags)
			continue
		}

		dbClients[clients.Region(provider.ID)] = dbClient
	}

	// Check if at least one region has been initialized successfully
	if len(dbClients) == 0 {
		resp.Diagnostics.AddError("Initialization Error", "No database regions were initialized. Please check your configuration.")
		return
	}

	log.Printf("[DEBUG] Initialized DB clients for regions: %v\n", dbClients)

	// Store the configured values in the provider instance for later use
	p.client = &utils.ProviderMeta{
		DB:             dbClients,
		Frontegg:       fronteggClient,
		CloudAPI:       cloudAPIClient,
		DefaultRegion:  clients.Region(defaultRegion),
		RegionsEnabled: regionsEnabled,
	}
}
