package resources

import (
	"context"
	"database/sql"
	"errors"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func sourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s, diags := sourceReadParams(ctx, d, meta)
	if s == nil {
		return diags
	}

	// Set the webhook URL only if this is a webhook source
	if s.WebhookUrl.Valid {
		_ = d.Set("url", s.WebhookUrl.String)
	}

	return nil
}

// sourceReadParams populates the fields every source shares and hands back the
// scanned row so a source type can set its own fields without a second query.
// A nil result with nil diagnostics means the source no longer exists.
func sourceReadParams(ctx context.Context, d *schema.ResourceData, meta interface{}) (*materialize.SourceParams, diag.Diagnostics) {
	i := d.Id()

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	// A source's size is backed by its linked cluster. As of Materialize
	// v26.34 a resize is graceful and proceeds in the background, so
	// mz_clusters still reports the old size until it commits. Read the
	// in-flight target before the source row, as clusterRead does, so a resize
	// that finalizes between the two reads is not missed by both.
	pending, inFlight, err := materialize.ScanSourcePendingReconfiguration(metaDb, utils.ExtractId(i))
	if err != nil {
		return nil, diag.FromErr(err)
	}

	s, err := materialize.ScanSource(metaDb, utils.ExtractId(i))
	if errors.Is(err, sql.ErrNoRows) {
		d.SetId("")
		return nil, nil
	} else if err != nil {
		return nil, diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

	if err := d.Set("name", s.SourceName.String); err != nil {
		return nil, diag.FromErr(err)
	}

	if err := d.Set("schema_name", s.SchemaName.String); err != nil {
		return nil, diag.FromErr(err)
	}

	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return nil, diag.FromErr(err)
	}

	size := s.Size.String
	if inFlight && pending.Size.Valid {
		size = pending.Size.String
	}
	if err := d.Set("size", size); err != nil {
		return nil, diag.FromErr(err)
	}

	if err := d.Set("cluster_name", s.ClusterName.String); err != nil {
		return nil, diag.FromErr(err)
	}

	if err := d.Set("ownership_role", s.OwnerName.String); err != nil {
		return nil, diag.FromErr(err)
	}

	b := materialize.Source{SourceName: s.SourceName.String, SchemaName: s.SchemaName.String, DatabaseName: s.DatabaseName.String}
	if err := d.Set("qualified_sql_name", b.QualifiedName()); err != nil {
		return nil, diag.FromErr(err)
	}

	if err := d.Set("comment", s.Comment.String); err != nil {
		return nil, diag.FromErr(err)
	}

	return &s, nil
}

// setSourceConnection sets the connection block of a source that reads through
// a connection, such as kafka_connection or postgres_connection.
func setSourceConnection(d *schema.ResourceData, key string, s *materialize.SourceParams) error {
	if !s.ConnectionName.Valid || !s.ConnectionSchemaName.Valid || !s.ConnectionDatabaseName.Valid {
		return nil
	}
	return d.Set(key, []interface{}{
		map[string]interface{}{
			"name":          s.ConnectionName.String,
			"schema_name":   s.ConnectionSchemaName.String,
			"database_name": s.ConnectionDatabaseName.String,
		},
	})
}

func sourceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if diags := alterSource(d, meta); diags != nil {
		return diags
	}
	return sourceRead(ctx, d, meta)
}

// alterSource applies the changes every source type supports in place: name,
// ownership_role and comment.
func alterSource(d *schema.ResourceData, meta any) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: materialize.BaseSource, Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: materialize.BaseSource, Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewSource(metaDb, o)
		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
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

	return nil
}

func sourceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSource(metaDb, o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
