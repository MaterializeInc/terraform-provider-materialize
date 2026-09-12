package clients

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/jmoiron/sqlx"
)

type DBClient struct {
	*sqlx.DB
}

func NewDBClient(host, user, password string, port int, database, application_name, version, sslmode string, options map[string]string) (*DBClient, diag.Diagnostics) {
	var diags diag.Diagnostics

	if application_name == "" {
		application_name = fmt.Sprintf("terraform-provider-materialize v%s", version)
	}

	connStr := buildConnectionString(host, user, password, port, database, sslmode, application_name, options)
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

func buildConnectionString(host, user, password string, port int, database, sslmode, application_name string, options map[string]string) string {
	parts := []string{`--transaction_isolation=strict\ serializable`}

	keys := make([]string, 0, len(options))
	for k := range options {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("--%s=%s", escapeOptionToken(k), escapeOptionToken(options[k])))
	}

	query := url.Values{
		"application_name": {application_name},
		"sslmode":          {sslmode},
		"options":          {strings.Join(parts, " ")},
	}.Encode()

	// Encode writes spaces as "+", which is an HTML form convention rather than
	// RFC 3986. pgx v5 does not read it back as a space, so the escaped spaces in
	// the options string would reach the server as literal "+" and the whole
	// option would be ignored. Every "+" left by Encode is a space, since a
	// literal one comes back as %2B.
	query = strings.ReplaceAll(query, "+", "%20")

	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     fmt.Sprintf("%s:%d", host, port),
		Path:     database,
		RawQuery: query,
	}

	return u.String()
}

func escapeOptionToken(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, ` `, `\ `)
	return s
}

func (c *DBClient) SQLX() *sqlx.DB {
	return c.DB
}
