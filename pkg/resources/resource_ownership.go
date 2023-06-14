package resources

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var ownershipSchema = map[string]*schema.Schema{
	"object": {
		Description: "The object to manage ownership of.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Description: "The name of the object.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"schema_name": {
					Description: "The schema name of the object (if applicable).",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"database_name": {
					Description: "The database name of the object (if applicable).",
					Type:        schema.TypeString,
					Optional:    true,
				},
			},
		},
		Required: true,
		MinItems: 1,
		MaxItems: 1,
		ForceNew: true,
	},
	"object_qualified_sql_name": QualifiedNameSchema("object"),
	"object_type": {
		Description: "The type of object.",
		Type:        schema.TypeString,
		Required:    true,
		ValidateFunc: func(val any, key string) (warns []string, errs []error) {
			v := val.(string)

			objects := make([]string, len(materialize.ObjectPermissions))
			i := 0
			for k := range materialize.ObjectPermissions {
				objects[i] = k
				i++
			}

			for _, b := range objects {
				if b == v {
					return
				}
			}

			errs = append(errs, fmt.Errorf("[ERROR] %s is not of allowed object type: %s", v, objects))
			return
		}},
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
	i := d.Id()

	s, err := materialize.ScanOwnership(meta.(*sqlx.DB), i, objectType)
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("role_name", s.RoleName.String); err != nil {
		return diag.FromErr(err)
	}

	if v, ok := d.GetOk("object"); ok {
		o := materialize.GetObjectSchemaStruct(v)
		if err := d.Set("object_qualified_sql_name", o.QualifiedName()); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func ownershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	objectType := d.Get("object_type").(string)

	b := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), objectType)

	if v, ok := d.GetOk("role_name"); ok {
		b.RoleName(v.(string))
	}

	if v, ok := d.GetOk("object"); ok {
		o := materialize.GetObjectSchemaStruct(v)
		b.Object(o)
	}

	// create resource as ALTER
	if err := b.Alter(); err != nil {
		return diag.FromErr(err)
	}

	// set id
	var o = materialize.GetObjectSchemaStruct(d.Get("object"))
	i, err := materialize.OwnershipId(meta.(*sqlx.DB), objectType, o.Name, o.SchemaName, o.DatabaseName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return ownershipRead(ctx, d, meta)
}

func ownershipUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	objectType := d.Get("object_type").(string)

	b := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), objectType)

	object := materialize.GetObjectSchemaStruct(d.Get("object"))
	b.Object(object)

	if d.HasChange("role_name") {
		_, newRole := d.GetChange("role_name")
		b.RoleName(newRole.(string))
	}

	return ownershipRead(ctx, d, meta)
}

func ownershipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// ownership cannot be removed, rather remove the resource from state
	d.SetId("")
	return nil
}
