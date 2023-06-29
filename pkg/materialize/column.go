package materialize

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type ColumnParams struct {
	Id            sql.NullString `db:"id"`
	Name          sql.NullString `db:"name"`
	Position      sql.NullString `db:"position"`
	Nullable      sql.NullBool   `db:"nullable"`
	Type          sql.NullString `db:"type"`
	Default       sql.NullString `db:"default"`
	IndexedColumn sql.NullBool   `db:"indexed_column"`
	IndexName     sql.NullString `db:"index_name"`
	IndexId       sql.NullString `db:"index_id"`
}

var columnQuery = NewBaseQuery(`
	SELECT
		mz_columns.id,
		mz_columns.name,
		mz_columns.position,
		mz_columns.nullable,
		mz_columns.type,
		mz_columns.default,
		CASE WHEN mz_index_columns.index_id IS NOT NULL THEN true ELSE false END AS indexed_column,
		mz_indexes.name AS index_name,
		mz_indexes.id AS index_id
	FROM mz_columns
	LEFT JOIN mz_indexes
		ON mz_columns.id = mz_indexes.on_id
	LEFT JOIN mz_index_columns
		ON mz_index_columns.index_id = mz_indexes.id
		AND mz_index_columns.index_position = mz_columns.position`).Order("mz_columns.position")

func ListColumns(conn *sqlx.DB, objectId string) ([]ColumnParams, error) {
	p := map[string]string{"mz_columns.id": objectId}
	q := columnQuery.QueryPredicate(p)

	var c []ColumnParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
