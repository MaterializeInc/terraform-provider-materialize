package provider

import (
	"context"
	"fmt"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/datasources"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/resources"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func Provider(version string) *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Materialize host. Can also come from the `MZ_PASSWORD` environment variable.",
				DefaultFunc: schema.EnvDefaultFunc("MZ_PASSWORD", nil),
			},
			"database": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Materialize database. Can also come from the `MZ_DATABASE` environment variable. Defaults to `materialize`.",
				DefaultFunc: schema.EnvDefaultFunc("MZ_DATABASE", "materialize"),
			},
			"sslmode": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MZ_SSLMODE", "require"),
				Description: "For testing purposes, the SSL mode to use.",
			},
			// TODO: Switch name to Admin Endpoint for consistency:
			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MZ_ENDPOINT", "https://admin.cloud.materialize.com"),
				Description: "The endpoint for the Materialize API.",
			},
			"cloud_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MZ_CLOUD_ENDPOINT", "https://api.cloud.materialize.com"),
				Description: "The endpoint for the Materialize Cloud API.",
			},
			"base_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MZ_BASE_ENDPOINT", "https://cloud.materialize.com"),
				Description: "The base endpoint for Materialize.",
			},
			"default_region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The default region if not specified in the resource",
				DefaultFunc: schema.EnvDefaultFunc("MZ_DEFAULT_REGION", "aws/us-east-1"),
			},
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Materialize host. Can also come from the `MZ_HOST` environment variable.",
				DefaultFunc: schema.EnvDefaultFunc("MZ_HOST", nil),
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The Materialize SQL port. Can also come from the `MZ_PORT` environment variable.",
				DefaultFunc: schema.EnvDefaultFunc("MZ_PORT", 6875),
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Materialize username. Can also come from the `MZ_USERNAME` environment variable.",
				DefaultFunc: schema.EnvDefaultFunc("MZ_USERNAME", "materialize"),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"materialize_app_password":                         resources.AppPassword(),
			"materialize_user":                                 resources.User(),
			"materialize_cluster":                              resources.Cluster(),
			"materialize_cluster_grant":                        resources.GrantCluster(),
			"materialize_cluster_grant_default_privilege":      resources.GrantClusterDefaultPrivilege(),
			"materialize_cluster_replica":                      resources.ClusterReplica(),
			"materialize_connection_aws":                       resources.ConnectionAws(),
			"materialize_connection_aws_privatelink":           resources.ConnectionAwsPrivatelink(),
			"materialize_connection_confluent_schema_registry": resources.ConnectionConfluentSchemaRegistry(),
			"materialize_connection_kafka":                     resources.ConnectionKafka(),
			"materialize_connection_mysql":                     resources.ConnectionMySQL(),
			"materialize_connection_postgres":                  resources.ConnectionPostgres(),
			"materialize_connection_ssh_tunnel":                resources.ConnectionSshTunnel(),
			"materialize_connection_grant":                     resources.GrantConnection(),
			"materialize_connection_grant_default_privilege":   resources.GrantConnectionDefaultPrivilege(),
			"materialize_database":                             resources.Database(),
			"materialize_database_grant":                       resources.GrantDatabase(),
			"materialize_database_grant_default_privilege":     resources.GrantDatabaseDefaultPrivilege(),
			"materialize_grant_system_privilege":               resources.GrantSystemPrivilege(),
			"materialize_index":                                resources.Index(),
			"materialize_materialized_view":                    resources.MaterializedView(),
			"materialize_materialized_view_grant":              resources.GrantMaterializedView(),
			"materialize_network_policy":                       resources.NetworkPolicy(),
			"materialize_region":                               resources.Region(),
			"materialize_role":                                 resources.Role(),
			"materialize_role_grant":                           resources.GrantRole(),
			"materialize_role_parameter":                       resources.RoleParameter(),
			"materialize_schema":                               resources.Schema(),
			"materialize_scim_config":                          resources.SCIM2Configuration(),
			"materialize_scim_group":                           resources.SCIM2Group(),
			"materialize_scim_group_users":                     resources.SCIM2GroupUsers(),
			"materialize_scim_group_roles":                     resources.SCIM2GroupRoles(),
			"materialize_sso_config":                           resources.SSOConfiguration(),
			"materialize_sso_domain":                           resources.SSODomain(),
			"materialize_sso_group_mapping":                    resources.SSORoleGroupMapping(),
			"materialize_sso_default_roles":                    resources.SSODefaultRoles(),
			"materialize_schema_grant":                         resources.GrantSchema(),
			"materialize_schema_grant_default_privilege":       resources.GrantSchemaDefaultPrivilege(),
			"materialize_secret":                               resources.Secret(),
			"materialize_secret_grant":                         resources.GrantSecret(),
			"materialize_secret_grant_default_privilege":       resources.GrantSecretDefaultPrivilege(),
			"materialize_sink_kafka":                           resources.SinkKafka(),
			"materialize_source_kafka":                         resources.SourceKafka(),
			"materialize_source_load_generator":                resources.SourceLoadgen(),
			"materialize_source_mysql":                         resources.SourceMySQL(),
			"materialize_source_postgres":                      resources.SourcePostgres(),
			"materialize_source_webhook":                       resources.SourceWebhook(),
			"materialize_source_grant":                         resources.GrantSource(),
			"materialize_system_parameter":                     resources.SystemParameter(),
			"materialize_table":                                resources.Table(),
			"materialize_source_table_kafka":                   resources.SourceTableKafka(),
			"materialize_source_table_load_generator":          resources.SourceTableLoadGen(),
			"materialize_source_table_mysql":                   resources.SourceTableMySQL(),
			"materialize_source_table_postgres":                resources.SourceTablePostgres(),
			"materialize_source_table_webhook":                 resources.SourceTableWebhook(),
			"materialize_table_grant":                          resources.GrantTable(),
			"materialize_table_grant_default_privilege":        resources.GrantTableDefaultPrivilege(),
			"materialize_type":                                 resources.Type(),
			"materialize_type_grant":                           resources.GrantType(),
			"materialize_type_grant_default_privilege":         resources.GrantTypeDefaultPrivilege(),
			"materialize_view":                                 resources.View(),
			"materialize_view_grant":                           resources.GrantView(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"materialize_cluster":           datasources.Cluster(),
			"materialize_cluster_replica":   datasources.ClusterReplica(),
			"materialize_connection":        datasources.Connection(),
			"materialize_current_database":  datasources.CurrentDatabase(),
			"materialize_current_cluster":   datasources.CurrentCluster(),
			"materialize_database":          datasources.Database(),
			"materialize_egress_ips":        datasources.EgressIps(),
			"materialize_index":             datasources.Index(),
			"materialize_materialized_view": datasources.MaterializedView(),
			"materialize_network_policy":    datasources.NetworkPolicy(),
			"materialize_region":            datasources.Region(),
			"materialize_role":              datasources.Role(),
			"materialize_schema":            datasources.Schema(),
			"materialize_secret":            datasources.Secret(),
			"materialize_sink":              datasources.Sink(),
			"materialize_source":            datasources.Source(),
			"materialize_source_reference":  datasources.SourceReference(),
			"materialize_source_table":      datasources.SourceTable(),
			"materialize_scim_groups":       datasources.SCIMGroups(),
			"materialize_scim_configs":      datasources.SCIMConfigs(),
			"materialize_sso_config":        datasources.SSOConfig(),
			"materialize_system_parameter":  datasources.SystemParameter(),
			"materialize_table":             datasources.Table(),
			"materialize_type":              datasources.Type(),
			"materialize_user":              datasources.User(),
			"materialize_view":              datasources.View(),
		},
		ConfigureContextFunc: func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
			return providerConfigure(ctx, d, version)
		},
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData, version string) (interface{}, diag.Diagnostics) {
	// Check for self-hosted configuration
	if host := d.Get("host").(string); host != "" {
		log.Printf("[DEBUG] Configuring self-hosted provider")
		return configureSelfHosted(ctx, d, version)
	}
	log.Printf("[DEBUG] Configuring SaaS provider")
	return configureSaaS(ctx, d, version)
}

func configureSelfHosted(ctx context.Context, d *schema.ResourceData, version string) (interface{}, diag.Diagnostics) {
	host := d.Get("host").(string)
	port := d.Get("port").(int)
	database := d.Get("database").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	sslmode := d.Get("sslmode").(string)
	application_name := fmt.Sprintf("terraform-provider-materialize v%s", version)

	// Initialize single DB client for self-hosted
	dbClient, diags := clients.NewDBClient(
		host,
		username,
		password,
		port,
		database,
		application_name,
		version,
		sslmode,
	)
	if diags.HasError() {
		return nil, diags
	}

	// Create provider meta for self-hosted
	dbClients := make(map[clients.Region]*clients.DBClient)
	dbClients["self-hosted"] = dbClient

	providerMeta := &utils.ProviderMeta{
		Mode:          utils.ModeSelfHosted,
		DB:            dbClients,
		DefaultRegion: "self-hosted",
		RegionsEnabled: map[clients.Region]bool{
			"self-hosted": true,
		},
	}

	return providerMeta, nil
}

func configureSaaS(ctx context.Context, d *schema.ResourceData, version string) (interface{}, diag.Diagnostics) {
	password := d.Get("password").(string)
	database := d.Get("database").(string)
	sslmode := d.Get("sslmode").(string)
	endpoint := d.Get("endpoint").(string)
	cloudEndpoint := d.Get("cloud_endpoint").(string)
	defaultRegion := clients.Region(d.Get("default_region").(string))
	baseEndpoint := d.Get("base_endpoint").(string)
	application_name := fmt.Sprintf("terraform-provider-materialize v%s", version)

	err := utils.SetDefaultRegion(string(defaultRegion))
	if err != nil {
		return nil, diag.FromErr(err)
	}

	// Initialize the Frontegg client
	fronteggClient, err := clients.NewFronteggClient(ctx, password, endpoint)
	if err != nil {
		return nil, diag.Errorf("Unable to create Frontegg client: %s", err)
	}

	// Initialize the Cloud API client using the Frontegg client and endpoint
	cloudAPIClient := clients.NewCloudAPIClient(fronteggClient, cloudEndpoint, baseEndpoint)
	regionsEnabled := make(map[clients.Region]bool)

	// Get the list of cloud providers
	providers, err := cloudAPIClient.ListCloudProviders(ctx)
	if err != nil {
		return nil, diag.Errorf("Unable to list cloud providers: %s", err)
	}

	// Store the DB clients for all regions.
	dbClients := make(map[clients.Region]*clients.DBClient)

	for _, provider := range providers {
		regionDetails, err := cloudAPIClient.GetRegionDetails(ctx, provider)

		log.Printf("[DEBUG] Region details for provider %s: %v\n", provider.ID, regionDetails)

		if err != nil {
			log.Printf("[ERROR] Error getting region details for provider %s: %v\n", provider.ID, err)
			continue
		}

		// Check if regionDetails or RegionInfo is nil before proceeding
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
		dbClient, diags := clients.NewDBClient(host, user, password, port, database, application_name, version, sslmode)
		if diags.HasError() {
			log.Printf("[ERROR] Error initializing DB client for region %s: %v\n", provider.ID, diags)
			continue
		}

		dbClients[clients.Region(provider.ID)] = dbClient
	}

	// Check if at least one region has been initialized successfully
	if len(dbClients) == 0 {
		return nil, diag.Errorf("No database regions were initialized. Please check your configuration.")
	}

	log.Printf("[DEBUG] Initialized DB clients for regions: %v\n", dbClients)

	// Fetch Frontegg roles and store them in the provider meta
	fronteggRoles, err := frontegg.ListFronteggRoles(ctx, fronteggClient)
	if err != nil {
		return nil, diag.Errorf("Unable to fetch Frontegg roles: %s", err)
	}

	// Construct and return the provider meta with all clients initialized.
	providerMeta := &utils.ProviderMeta{
		Mode:           utils.ModeSaaS,
		DB:             dbClients,
		Frontegg:       fronteggClient,
		CloudAPI:       cloudAPIClient,
		DefaultRegion:  clients.Region(defaultRegion),
		RegionsEnabled: regionsEnabled,
		FronteggRoles:  fronteggRoles,
	}

	return providerMeta, nil
}
