package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/datasources"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/resources"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Materialize host. Can also come from the `MZ_HOST` environment variable.",
				DefaultFunc: schema.EnvDefaultFunc("MZ_HOST", nil),
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Materialize username. Can also come from the `MZ_USER` environment variable.",
				DefaultFunc: schema.EnvDefaultFunc("MZ_USER", nil),
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Materialize host. Can also come from the `MZ_PW` environment variable.",
				DefaultFunc: schema.EnvDefaultFunc("MZ_PW", nil),
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The Materialize port number to connect to at the server host. Can also come from the `MZ_PORT` environment variable. Defaults to 6875.",
				DefaultFunc: schema.EnvDefaultFunc("MZ_PORT", 6875),
			},
			"database": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Materialize database. Can also come from the `MZ_DATABASE` environment variable. Defaults to `materialize`.",
				DefaultFunc: schema.EnvDefaultFunc("MZ_DATABASE", "materialize"),
			},
			"application_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "terraform-provider-materialize",
				Description: "The application name to include in the connection string",
			},
			"testing": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MZ_TESTING", false),
				Description: "Enable to test the provider locally",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"materialize_cluster":                              resources.Cluster(),
			"materialize_cluster_replica":                      resources.ClusterReplica(),
			"materialize_connection_aws_privatelink":           resources.ConnectionAwsPrivatelink(),
			"materialize_connection_confluent_schema_registry": resources.ConnectionConfluentSchemaRegistry(),
			"materialize_connection_kafka":                     resources.ConnectionKafka(),
			"materialize_connection_postgres":                  resources.ConnectionPostgres(),
			"materialize_connection_ssh_tunnel":                resources.ConnectionSshTunnel(),
			"materialize_database":                             resources.Database(),
			"materialize_grant_cluster":                        resources.GrantCluster(),
			"materialize_grant_connection":                     resources.GrantConnection(),
			"materialize_grant_database":                       resources.GrantDatabase(),
			"materialize_grant_materialized_view":              resources.GrantMaterializedView(),
			"materialize_grant_schema":                         resources.GrantSchema(),
			"materialize_grant_secret":                         resources.GrantSecret(),
			"materialize_grant_source":                         resources.GrantSource(),
			"materialize_grant_table":                          resources.GrantTable(),
			"materialize_grant_type":                           resources.GrantType(),
			"materialize_grant_view":                           resources.GrantView(),
			"materialize_index":                                resources.Index(),
			"materialize_materialized_view":                    resources.MaterializedView(),
			"materialize_role":                                 resources.Role(),
			"materialize_schema":                               resources.Schema(),
			"materialize_secret":                               resources.Secret(),
			"materialize_sink_kafka":                           resources.SinkKafka(),
			"materialize_source_kafka":                         resources.SourceKafka(),
			"materialize_source_load_generator":                resources.SourceLoadgen(),
			"materialize_source_postgres":                      resources.SourcePostgres(),
			"materialize_table":                                resources.Table(),
			"materialize_type":                                 resources.Type(),
			"materialize_view":                                 resources.View(),
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
			"materialize_role":              datasources.Role(),
			"materialize_schema":            datasources.Schema(),
			"materialize_secret":            datasources.Secret(),
			"materialize_sink":              datasources.Sink(),
			"materialize_source":            datasources.Source(),
			"materialize_table":             datasources.Table(),
			"materialize_type":              datasources.Type(),
			"materialize_view":              datasources.View(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func connectionString(host, username, password string, port int, database string, testing bool, application string) string {
	c := strings.Builder{}
	c.WriteString(fmt.Sprintf("postgres://%s:%s@%s:%d/%s", username, password, host, port, database))

	params := []string{}

	if testing {
		params = append(params, "sslmode=disable")
		params = append(params, "enable_rbac_checks=true")
		params = append(params, "enable_ld_rbac_checks=true")
		params = append(params, "enable_managed_clusters=true")
	} else {
		params = append(params, "sslmode=require")
	}

	params = append(params, fmt.Sprintf("application_name=%s", application))
	p := strings.Join(params[:], "&")

	c.WriteString(fmt.Sprintf("?%s", p))
	return c.String()
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	host := d.Get("host").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	port := d.Get("port").(int)
	database := d.Get("database").(string)
	application := d.Get("application_name").(string)
	testing := d.Get("testing").(bool)

	connStr := connectionString(host, username, password, port, database, testing, application)

	var diags diag.Diagnostics
	db, err := sqlx.Open("pgx", connStr)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Materialize client",
			Detail:   "Unable to authenticate user for authenticated Materialize client",
		})
		return nil, diags
	}

	return db, diags
}
