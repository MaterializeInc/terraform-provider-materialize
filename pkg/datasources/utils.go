package datasources

import (
	"fmt"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SetId(region, resource, databaseName, schemaName string, d *schema.ResourceData) {
	var id string
	if databaseName != "" && schemaName != "" {
		id = fmt.Sprintf("%s|%s|%s", databaseName, schemaName, resource)
	} else if databaseName != "" {
		id = fmt.Sprintf("%s|%s", databaseName, resource)
	} else {
		id = resource
	}

	d.SetId(utils.TransformIdWithRegion(region, id))
}

func RegionSchema() *schema.Schema {
	return &schema.Schema{
		Description: "The region in which the resource is located.",
		Type:        schema.TypeString,
		Computed:    true,
	}
}
