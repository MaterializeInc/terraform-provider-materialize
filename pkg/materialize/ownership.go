package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type OwnershipBuilder struct {
	ddl        Builder
	objectType string
	object     ObjectSchemaStruct
	roleName   string
}

func NewOwnershipBuilder(conn *sqlx.DB, objectType string) *OwnershipBuilder {
	return &OwnershipBuilder{
		ddl:        Builder{conn, Ownership},
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
	return b.ddl.exec(q)
}

type OwnershipParams struct {
	ObjectId    sql.NullString `db:"id"`
	ObjectType  sql.NullString `db:"obj_type"`
	OwnershipId sql.NullString `db:"owner_id"`
	RoleName    sql.NullString `db:"role_name"`
}

var ownershipQuery = NewBaseQuery(`
	SELECT
		mz_objects.id,
		mz_objects.type AS obj_type,
		mz_objects.owner_id AS schema_name,
		mz_roles.name AS role_name
	FROM mz_objects
	JOIN mz_roles
		ON mz_objects.owner_id = mz_roles.id
	LEFT JOIN mz_schemas
		ON mz_objects.schema_id = mz_schemas.id
	LEFT JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id;`)

func OwnershipId(conn *sqlx.DB, objectType, objectName, schemaName, databaseName string) (string, error) {
	p := map[string]string{
		"mz_objects.type":  objectType,
		"mz_objects.name":  objectName,
		"mz_schemas.name":  schemaName,
		"mz_database.name": databaseName,
	}
	q := ownershipQuery.QueryPredicate(p)

	var c OwnershipParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	return OwnershipResourceId(objectType, c.OwnershipId.String), nil
}

func ScanOwnership(conn *sqlx.DB, id, objectType string) (OwnershipParams, error) {
	p := map[string]string{
		"mz_objects.type": objectType,
		"mz_objects.id":   id,
	}

	q := ownershipQuery.QueryPredicate(p)

	var c OwnershipParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
