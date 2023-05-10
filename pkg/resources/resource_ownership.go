package resources

import (
	"context"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var allowedOwners = []string{
	"CLUSTER",
	"CLUSTER_REPLICA",
	"CONNECTION",
	"DATABASE",
	"SCHEMA",
	"SOURCE",
	"SINK",
	"VIEW",
	"MATERIALIZED VIEW",
	"TABLE",
	"TYPE",
	"SECRET",
}

var ownershipSchema = map[string]*schema.Schema{
	"object":                    IdentifierSchema("object", "The identifier of the item you want to set ownership.", true),
	"object_qualified_sql_name": QualifiedNameSchema("object"),
	"object_type": {
		Description:  "The type of object.",
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice(allowedOwners, true),
	},
	"role_name": {
		Description: "The role to assoicate as the owner of the object.",
		Type:        schema.TypeString,
		Required:    true,
	},
}

func Ownership() *schema.Resource {
	return &schema.Resource{
		Description: "The owner of an item in Materialize.",

		CreateContext: ownershipCreate,
		ReadContext:   ownershipRead,
		UpdateContext: ownershipUpdate,
		DeleteContext: ownershipDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: ownershipSchema,
	}
}

func ownershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	objectType := d.Get("object_type").(string)
	id := d.Id()

	builder := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), objectType)

	catalogId := materialize.OwnershipCatalogId(id)
	params, err := builder.Params(catalogId)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("role_name", params.RoleName); err != nil {
		return diag.FromErr(err)
	}

	qn := d.Get("object").(materialize.IdentifierSchemaStruct)
	if err := d.Set("qualified_sql_name", qn.QualifiedName()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ownershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	objectType := d.Get("object_type").(string)

	builder := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), objectType)

	if v, ok := d.GetOk("role_name"); ok {
		builder.RoleName(v.(string))
	}

	if v, ok := d.GetOk("object"); ok {
		builder.Object(v.(materialize.IdentifierSchemaStruct))
	}

	if err := builder.Alter(); err != nil {
		return diag.FromErr(err)
	}

	return ownershipRead(ctx, d, meta)
}

func ownershipUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	object := d.Get("object").(materialize.IdentifierSchemaStruct)
	objectType := d.Get("object_type").(string)
	b := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), objectType)

	b.Object(object)

	if d.HasChange("role_name") {
		_, newRole := d.GetChange("role_name")

		b.RoleName(newRole.(string))

		if err := b.Alter(); err != nil {
			log.Printf("[ERROR] updating ownership of %v", d.Id())
			return diag.FromErr(err)
		}
	}

	return ownershipRead(ctx, d, meta)
}

func ownershipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// ownership cannot be removed, rather remove the resource from state
	d.SetId("")
	return nil
}
