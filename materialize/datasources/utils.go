package datasources

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SetId(resource, databaseName, schemaName string, d *schema.ResourceData) {
	var id string
	if databaseName != "" && schemaName != "" {
		id = fmt.Sprintf("%s|%s|%s", databaseName, schemaName, resource)
	} else if databaseName != "" {
		id = fmt.Sprintf("%s|%s", databaseName, resource)
	} else {
		id = resource
	}

	d.SetId(id)
}
