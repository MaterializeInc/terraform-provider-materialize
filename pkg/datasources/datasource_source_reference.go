package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SourceReference() *schema.Resource {
	return &schema.Resource{
		ReadContext: sourceReferenceRead,
		Description: "The `materialize_source_reference` data source retrieves a list of *available* upstream references for a given Materialize source. These references represent potential tables that can be created based on the source, but they do not necessarily indicate references the source is already ingesting. This allows users to see all upstream data that could be materialized into tables.",
		Schema: map[string]*schema.Schema{
			"source_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the source to get references for",
			},
			"references": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The source references",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The namespace of the reference",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the reference",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The last update timestamp of the reference",
						},
						"columns": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The columns of the reference",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"source_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the source",
						},
						"source_schema_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The schema name of the source",
						},
						"source_database_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The database name of the source",
						},
						"source_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of the source",
						},
					},
				},
			},
			"region": RegionSchema(),
		},
	}
}

func sourceReferenceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sourceID := d.Get("source_id").(string)
	sourceID = utils.ExtractId(sourceID)

	var diags diag.Diagnostics

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	sourceReference, err := materialize.ListSourceReferences(metaDb, sourceID)
	if err != nil {
		return diag.FromErr(err)
	}

	referenceFormats := []map[string]interface{}{}
	for _, sr := range sourceReference {
		referenceMap := map[string]interface{}{
			"namespace":            sr.Namespace.String,
			"name":                 sr.Name.String,
			"updated_at":           sr.UpdatedAt.String,
			"columns":              sr.Columns,
			"source_name":          sr.SourceName.String,
			"source_schema_name":   sr.SourceSchemaName.String,
			"source_database_name": sr.SourceDBName.String,
			"source_type":          sr.SourceType.String,
		}
		referenceFormats = append(referenceFormats, referenceMap)
	}

	if err := d.Set("references", referenceFormats); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), sourceID))

	return diags
}
