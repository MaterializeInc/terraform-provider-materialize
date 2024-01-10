package clients

import (
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/jmoiron/sqlx"
)

type DBClient struct {
	*sqlx.DB
}

func NewDBClient(host, user, password string, port int, database, application_name_suffix, version, sslmode string) (*DBClient, diag.Diagnostics) {
	var diags diag.Diagnostics

	application_name := fmt.Sprintf("terraform-provider-materialize v%s", version)
	if application_name_suffix != "" {
		application_name += " " + application_name_suffix
	}

	connStr := buildConnectionString(host, user, password, port, database, sslmode, application_name)
	db, err := sqlx.Open("pgx", connStr)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create database client",
			Detail:   fmt.Sprintf("Unable to authenticate user for database: %s", err),
		})
		return nil, diags
	}
	return &DBClient{DB: db}, diags
}

func buildConnectionString(host, user, password string, port int, database, sslmode, application_name string) string {
	url := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   database,
		RawQuery: url.Values{
			"application_name": {application_name},
			"sslmode":          {sslmode},
		}.Encode(),
	}

	return url.String()
}

func (c *DBClient) SQLX() *sqlx.DB {
	return c.DB
}
