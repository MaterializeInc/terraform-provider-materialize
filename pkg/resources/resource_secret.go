package resources

import (
	"context"
	"database/sql"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var secretSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("secret", true, false),
	"schema_name":        SchemaNameSchema("secret", false),
	"database_name":      DatabaseNameSchema("secret", false),
	"qualified_sql_name": QualifiedNameSchema("secret"),
	"comment":            CommentSchema(false),
	"value": {
		Description:  "The value for the secret. The value expression may not reference any relations, and must be a bytea string literal. Use value_wo for write-only ephemeral values that won't be stored in state.",
		Type:         schema.TypeString,
		Optional:     true,
		Sensitive:    true,
		ExactlyOneOf: []string{"value", "value_wo"},
	},
	"value_wo": {
		Description:  "Write-only value for the secret that supports ephemeral values and won't be stored in Terraform state or plan. The value expression may not reference any relations, and must be a bytea string literal. Requires Terraform 1.11+. Must be used with value_wo_version.",
		Type:         schema.TypeString,
		Optional:     true,
		Sensitive:    true,
		WriteOnly:    true,
		ExactlyOneOf: []string{"value", "value_wo"},
		RequiredWith: []string{"value_wo_version"},
	},
	"value_wo_version": {
		Description:  "Version number for the write-only value. Increment this to trigger an update of the secret value when using value_wo. Must be used with value_wo.",
		Type:         schema.TypeInt,
		Optional:     true,
		RequiredWith: []string{"value_wo"},
	},
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func Secret() *schema.Resource {
	return &schema.Resource{
		Description: "A secret securely stores sensitive credentials (like passwords and SSL keys) in Materialize’s secret management system.",

		CreateContext: secretCreate,
		ReadContext:   secretRead,
		UpdateContext: secretUpdate,
		DeleteContext: secretDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: secretSchema,
	}
}

func secretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	s, err := materialize.ScanSecret(metaDb, utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

	if err := d.Set("name", s.SecretName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", s.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ownership_role", s.OwnerName.String); err != nil {
		return diag.FromErr(err)
	}

	qn := materialize.QualifiedName(s.DatabaseName.String, s.SchemaName.String, s.SecretName.String)
	if err := d.Set("qualified_sql_name", qn); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func secretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	secretName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "SECRET", Name: secretName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSecretBuilder(metaDb, o)

	if v, ok := d.GetOk("value"); ok {
		b.Value(v.(string))
	} else if valueWo, _ := d.GetRawConfigAt(cty.GetAttrPath("value_wo")); !valueWo.IsNull() {
		if !valueWo.Type().Equals(cty.String) {
			return diag.Errorf("error retrieving write-only argument: value_wo - retrieved config value is not a string")
		}
		b.Value(valueWo.AsString())
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(metaDb, o)

		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// object comment
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(metaDb, o)

		if err := comment.Object(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// set id
	i, err := materialize.SecretId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return secretRead(ctx, d, meta)
}

func secretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	secretName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "SECRET", Name: secretName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSecretBuilder(metaDb, o)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: "SECRET", Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewSecretBuilder(metaDb, o)
		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("value") {
		_, newValue := d.GetChange("value")
		if err := b.UpdateValue(newValue.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("value_wo_version") {
		if valueWo, _ := d.GetRawConfigAt(cty.GetAttrPath("value_wo")); !valueWo.IsNull() {
			if !valueWo.Type().Equals(cty.String) {
				return diag.Errorf("error retrieving write-only argument: value_wo - retrieved config value is not a string")
			}
			if err := b.UpdateValue(valueWo.AsString()); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("ownership_role") {
		_, newRole := d.GetChange("ownership_role")
		b := materialize.NewOwnershipBuilder(metaDb, o)

		if err := b.Alter(newRole.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		b := materialize.NewCommentBuilder(metaDb, o)

		if err := b.Object(newComment.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return secretRead(ctx, d, meta)
}

func secretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	secretName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{Name: secretName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSecretBuilder(metaDb, o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
