package resources

import (
	"context"
	"database/sql"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var secretSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("secret", true, false),
	"schema_name":        SchemaNameSchema("secret", false),
	"database_name":      DatabaseNameSchema("secret", false),
	"qualified_sql_name": QualifiedNameSchema("secret"),
	"comment":            CommentSchema(false),
	"value": {
		Description: "The value for the secret. The value expression may not reference any relations, and must be a bytea string literal.",
		Type:        schema.TypeString,
		Required:    true,
		Sensitive:   true,
	},
	"ownership_role": OwnershipRoleSchema(),
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

	s, err := materialize.ScanSecret(meta.(*sqlx.DB), utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(i))

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

	o := materialize.MaterializeObject{ObjectType: "SECRET", Name: secretName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSecretBuilder(meta.(*sqlx.DB), o)

	if v, ok := d.GetOk("value"); ok {
		b.Value(v.(string))
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), o)

		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// object comment
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(meta.(*sqlx.DB), o)

		if err := comment.Object(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// set id
	i, err := materialize.SecretId(meta.(*sqlx.DB), o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(i))

	return secretRead(ctx, d, meta)
}

func secretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	secretName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	o := materialize.MaterializeObject{ObjectType: "SECRET", Name: secretName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSecretBuilder(meta.(*sqlx.DB), o)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: "SECRET", Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewSecretBuilder(meta.(*sqlx.DB), o)
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

	if d.HasChange("ownership_role") {
		_, newRole := d.GetChange("ownership_role")
		b := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), o)

		if err := b.Alter(newRole.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		b := materialize.NewCommentBuilder(meta.(*sqlx.DB), o)

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

	o := materialize.MaterializeObject{Name: secretName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSecretBuilder(meta.(*sqlx.DB), o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
