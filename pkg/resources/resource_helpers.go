package resources

import (
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

// Droppable is an interface for builders that support Drop operation
type Droppable interface {
	Drop() error
}

// applyOwnership applies ownership to a newly created resource.
// If the operation fails, it drops the resource and returns an error.
// This is a common pattern across connection, source, table, view, and other resources.
func applyOwnership(d *schema.ResourceData, metaDb *sqlx.DB, o materialize.MaterializeObject, builder Droppable) diag.Diagnostics {
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(metaDb, o)

		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			builder.Drop()
			return diag.FromErr(err)
		}
	}

	return nil
}

// applyComment applies a comment to a newly created resource.
// If the operation fails, it drops the resource and returns an error.
// This is a common pattern across connection, source, table, view, and other resources.
func applyComment(d *schema.ResourceData, metaDb *sqlx.DB, o materialize.MaterializeObject, builder Droppable) diag.Diagnostics {
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(metaDb, o)

		if err := comment.Object(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			builder.Drop()
			return diag.FromErr(err)
		}
	}

	return nil
}
