package materialize

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type TableColumnParams struct {
	Id       sql.NullString `db:"id"`
	Name     sql.NullString `db:"name"`
	Position sql.NullString `db:"position"`
	Nullable sql.NullBool   `db:"nullable"`
	Comment  sql.NullString `db:"comment"`
	Type     sql.NullString `db:"type"`
	Default  sql.NullString `db:"default"`
}

var tableColumnQuery = NewBaseQuery(`
	SELECT
		mz_columns.id,
		mz_columns.name,
		mz_columns.position,
		mz_columns.nullable,
		comments.comment,
		mz_columns.type,
		mz_columns.default
	FROM mz_columns
	LEFT JOIN (
		SELECT id, object_sub_id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'table'
	) comments
		ON mz_columns.id = comments.id
		AND mz_columns.position = comments.object_sub_id`).Order("mz_columns.position")

func ListTableColumns(conn *sqlx.DB, objectId string) ([]TableColumnParams, error) {
	p := map[string]string{"mz_columns.id": objectId}
	q := tableColumnQuery.QueryPredicate(p)

	var c []TableColumnParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}

type IndexColumnParams struct {
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

var indexColumnQuery = NewBaseQuery(`
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
		AND mz_index_columns.on_position = mz_columns.position`).Order("mz_columns.position")

func ListIndexColumns(conn *sqlx.DB, indexId string) ([]IndexColumnParams, error) {
	p := map[string]string{
		"mz_indexes.id": indexId,
	}
	q := indexColumnQuery.QueryPredicate(p)

	// Filter out non-indexed columns
	var allColumns []IndexColumnParams
	if err := conn.Select(&allColumns, q); err != nil {
		return allColumns, err
	}

	// Only keep columns that are part of the index
	var indexedColumns []IndexColumnParams
	for _, col := range allColumns {
		if col.IndexedColumn.Bool {
			indexedColumns = append(indexedColumns, col)
		}
	}

	return indexedColumns, nil
}
