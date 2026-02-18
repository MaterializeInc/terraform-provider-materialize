package resources

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var schemaSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("schema", true, false),
	"database_name":      DatabaseNameSchema("schema", false),
	"qualified_sql_name": QualifiedNameSchema("schema"),
	"comment":            CommentSchema(false),
	"ownership_role":     OwnershipRoleSchema(),
	"identify_by_name": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Use the schema name as the resource identifier in your state file, rather than the internal schema ID. Useful when schemas are recreated outside of Terraform (e.g. blue/green deployments), so the resource can be managed consistently when the ID changes.",
	},
	"region": RegionSchema(),
}

func Schema() *schema.Resource {
	return &schema.Resource{
		Description: "The second highest level namespace hierarchy in Materialize.",

		CreateContext: schemaCreate,
		ReadContext:   schemaRead,
		UpdateContext: schemaUpdate,
		DeleteContext: schemaDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schemaImport,
		},

		Schema: schemaSchema,
	}
}

func schemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fullId := d.Id()
	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	idType := utils.ExtractIdType(fullId)
	value := utils.ExtractId(fullId)
	useNameAsId := d.Get("identify_by_name").(bool)

	s, err := materialize.ScanSchema(metaDb, value, idType == "name")
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	if useNameAsId {
		d.SetId(utils.TransformIdWithTypeAndRegion(string(region), "name", s.DatabaseName.String+"|"+s.SchemaName.String))
	} else {
		d.SetId(utils.TransformIdWithRegion(string(region), s.SchemaId.String))
	}

	if err := d.Set("identify_by_name", useNameAsId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", s.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ownership_role", s.OwnerName.String); err != nil {
		return diag.FromErr(err)
	}
	qn := materialize.QualifiedName(s.DatabaseName.String, s.SchemaName.String)
	if err := d.Set("qualified_sql_name", qn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func schemaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "SCHEMA", Name: schemaName, DatabaseName: databaseName}
	b := materialize.NewSchemaBuilder(metaDb, o)

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if diags := applyOwnership(d, metaDb, o, b); diags != nil {
		return diags
	}

	// object comment
	if diags := applyComment(d, metaDb, o, b); diags != nil {
		return diags
	}

	// set id
	identifyByName := d.Get("identify_by_name").(bool)
	if identifyByName {
		d.SetId(utils.TransformIdWithTypeAndRegion(string(region), "name", databaseName+"|"+schemaName))
	} else {
		schemaId, err := materialize.SchemaId(metaDb, o)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(utils.TransformIdWithRegion(string(region), schemaId))
	}

	return schemaRead(ctx, d, meta)
}

func schemaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "SCHEMA", Name: schemaName, DatabaseName: databaseName}
	b := materialize.NewOwnershipBuilder(metaDb, o)

	if d.HasChange("identify_by_name") {
		_, newIdentifyByName := d.GetChange("identify_by_name")
		identifyByName := newIdentifyByName.(bool)

		fullId := d.Id()
		currentValue := utils.ExtractId(fullId)

		var newId string
		if identifyByName {
			newId = utils.TransformIdWithTypeAndRegion(string(region), "name", databaseName+"|"+schemaName)
		} else {
			schemaId, err := materialize.SchemaId(metaDb, o)
			if err != nil {
				return diag.FromErr(err)
			}
			newId = utils.TransformIdWithRegion(string(region), schemaId)
		}

		if currentValue != utils.ExtractId(newId) {
			d.SetId(newId)
		}
	}

	if d.HasChange("ownership_role") {
		_, newRole := d.GetChange("ownership_role")
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

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: "SCHEMA", Name: oldName.(string), DatabaseName: databaseName}
		b := materialize.NewSchemaBuilder(metaDb, o)
		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
		}

		// Update the ID after rename when using name-based identification,
		// otherwise schemaRead would look up the old name and fail.
		if d.Get("identify_by_name").(bool) {
			d.SetId(utils.TransformIdWithTypeAndRegion(string(region), "name", databaseName+"|"+newName.(string)))
		}
	}

	return schemaRead(ctx, d, meta)
}

func schemaImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return nil, err
	}

	fullId := d.Id()
	idType := utils.ExtractIdType(fullId)
	value := utils.ExtractId(fullId)
	identifyByName := idType == "name"

	s, err := materialize.ScanSchema(metaDb, value, identifyByName)
	if err != nil {
		return nil, fmt.Errorf("error importing schema %s: %w", fullId, err)
	}

	if identifyByName {
		d.SetId(utils.TransformIdWithTypeAndRegion(string(region), "name", s.DatabaseName.String+"|"+s.SchemaName.String))
	} else {
		d.SetId(utils.TransformIdWithRegion(string(region), s.SchemaId.String))
	}

	d.Set("identify_by_name", identifyByName)
	d.Set("name", s.SchemaName.String)
	d.Set("database_name", s.DatabaseName.String)
	d.Set("ownership_role", s.OwnerName.String)
	d.Set("comment", s.Comment.String)
	d.Set("qualified_sql_name", materialize.QualifiedName(s.DatabaseName.String, s.SchemaName.String))

	return []*schema.ResourceData{d}, nil
}

func schemaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{Name: schemaName, DatabaseName: databaseName}
	b := materialize.NewSchemaBuilder(metaDb, o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
