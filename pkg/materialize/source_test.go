package materialize

import (
	"reflect"
	"testing"
)

func TestAreEqual(t *testing.T) {
	o := areEqual(map[string]interface{}{"upstream_name": "upstream_name", "name": "name"}, map[string]interface{}{"upstream_name": "upstream_name", "name": "name"})
	if !o {
		t.Fatalf("Expected areEqual to be equal")
	}
}

func TestAreUnequal(t *testing.T) {
	o := areEqual(
		map[string]interface{}{"upstream_name": "upstream_name", "name": "name"},
		map[string]interface{}{"upstream_name": "diff", "name": "name"},
	)
	if o {
		t.Fatalf("Expected areEqual to not be equal")
	}
}

func TestDiffTableStructs(t *testing.T) {
	arr1 := []interface{}{
		map[string]interface{}{"upstream_name": "old", "upstream_schema_name": "public", "name": "old", "schema_name": "public"},
		map[string]interface{}{"upstream_name": "old_1", "upstream_schema_name": "public", "name": "old_2", "schema_name": "public"},
		map[string]interface{}{"upstream_name": "shared", "upstream_schema_name": "public", "name": "shared", "schema_name": "public"},
	}
	arr2 := []interface{}{
		map[string]interface{}{"upstream_name": "shared", "upstream_schema_name": "public", "name": "shared", "schema_name": "public"},
		map[string]interface{}{"upstream_name": "new", "upstream_schema_name": "public", "name": "new", "schema_name": "public"},
	}
	o := DiffTableStructs(arr1, arr2)
	e := []TableStruct{
		{UpstreamName: "old", UpstreamSchemaName: "public", Name: "old", SchemaName: "public"},
		{UpstreamName: "old_1", UpstreamSchemaName: "public", Name: "old_2", SchemaName: "public"},
	}
	if !reflect.DeepEqual(o, e) {
		t.Fatalf("Expect %s %s to be equal", o, e)
	}
}
