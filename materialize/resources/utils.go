package resources

import (
	"database/sql"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func ExecResource(conn *sql.DB, queryStr string) diag.Diagnostics {
	var diags diag.Diagnostics

	_, err := conn.Exec(queryStr)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
