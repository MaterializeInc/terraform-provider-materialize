package provider

import (
	"context"
	"fmt"

	"terraform-materialize/materialize/datasources"
	"terraform-materialize/materialize/resources"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" //PostgreSQL db
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Materialize host",
				DefaultFunc: schema.EnvDefaultFunc("MZ_HOST", nil),
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Materialize username",
				DefaultFunc: schema.EnvDefaultFunc("MZ_USER", nil),
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Materialize host",
				DefaultFunc: schema.EnvDefaultFunc("MZ_PW", nil),
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MZ_PORT", 6875),
				Description: "The Materialize port number to connect to at the server host",
			},
			"database": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MZ_DATABASE", "materialize"),
				Description: "The Materialize database",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"materialize_cluster":         resources.Cluster(),
			"materialize_cluster_replica": resources.ClusterReplica(),
			"materialize_connection":      resources.Connection(),
			"materialize_database":        resources.Database(),
			"materialize_schema":          resources.Schema(),
			"materialize_secret":          resources.Secret(),
			"materialize_sink":            resources.Sink(),
			"materialize_source":          resources.Source(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"materialize_cluster":          datasources.Cluster(),
			"materialize_cluster_replica":  datasources.ClusterReplica(),
			"materialize_connection":       datasources.Connection(),
			"materialize_current_database": datasources.CurrentDatabase(),
			"materialize_current_cluster":  datasources.CurrentCluster(),
			"materialize_database":         datasources.Database(),
			"materialize_schema":           datasources.Schema(),
			"materialize_secret":           datasources.Secret(),
			"materialize_sink":             datasources.Sink(),
			"materialize_source":           datasources.Source(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func connectionString(host string, username string, password string, port int, database string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=require", username, password, host, port, database)
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	host := d.Get("host").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	port := d.Get("port").(int)
	database := d.Get("database").(string)

	connStr := connectionString(host, username, password, port, database)

	var diags diag.Diagnostics
	db, err := sqlx.Open("postgres", connStr)
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
