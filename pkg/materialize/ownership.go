package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type OwnershipBuilder struct {
	conn       *sqlx.DB
	objectType string
	object     ObjectSchemaStruct
	roleName   string
}

func NewOwnershipBuilder(conn *sqlx.DB, objectType string) *OwnershipBuilder {
	return &OwnershipBuilder{
		conn:       conn,
		objectType: objectType,
	}
}

func (b *OwnershipBuilder) Object(o ObjectSchemaStruct) *OwnershipBuilder {
	b.object = o
	return b
}

func (b *OwnershipBuilder) RoleName(r string) *OwnershipBuilder {
	b.roleName = r
	return b
}

// generate a unique id `ownership|object_type|id` as there is no catalog id
func OwnershipResourceId(objectType, catalogId string) string {
	lo := strings.ToLower(objectType)
	fo := strings.ReplaceAll(lo, " ", "_")
	return fmt.Sprintf("ownership|%s|%s", fo, catalogId)
}

func OwnershipCatalogId(resourceId string) string {
	ci := strings.Split(resourceId, "|")
	return ci[len(ci)-1]
}

func (b *OwnershipBuilder) Alter() error {
	q := fmt.Sprintf(`ALTER %s %s OWNER TO %s;`, b.objectType, b.object.QualifiedName(), b.roleName)
	_, err := b.conn.Exec(q)

	if err != nil {
		return err
	}

	return nil
}

func (b *OwnershipBuilder) ReadId() (string, error) {
	o := ObjectPermissions[b.objectType]

	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`SELECT o.id FROM %s o`, o.CatalogTable))

	if b.object.SchemaName != "" {
		q.WriteString(` JOIN mz_schemas ON o.schema_id = mz_schemas.id`)
	}

	if b.object.DatabaseName != "" {
		q.WriteString(` JOIN mz_databases ON mz_schemas.database_id = mz_databases.id`)
	}

	// filter predicate
	q.WriteString(fmt.Sprintf(` WHERE o.name = %s`, QuoteString(b.object.Name)))

	if b.object.DatabaseName != "" {
		q.WriteString(fmt.Sprintf(`
		AND mz_databases.name = %s`, QuoteString(b.object.DatabaseName)))

		if b.object.SchemaName != "" {
			q.WriteString(fmt.Sprintf(` AND mz_schemas.name = %s`, QuoteString(b.object.SchemaName)))
		}
	}

	var i string
	if err := b.conn.QueryRowx(q.String()).Scan(&i); err != nil {
		return "", err
	}

	return OwnershipResourceId(b.objectType, i), nil
}

// return parameters specific to ownership
type OwnershipParams struct {
	OwnershipId sql.NullString `db:"owner_id"`
	RoleName    sql.NullString `db:"role_name"`
}

func (b *OwnershipBuilder) Params(catalogId string) (OwnershipParams, error) {
	o := ObjectPermissions[b.objectType]
	q := fmt.Sprintf(`
		SELECT
			o.owner_id,
			r.name AS role_name
		FROM %s o
		JOIN mz_roles r
			ON o.owner_id = r.id
		WHERE o.id = %s
	`, o.CatalogTable, QuoteString(catalogId))

	var s OwnershipParams
	if err := b.conn.Get(&s, q); err != nil {
		return s, err
	}
	return s, nil
}
