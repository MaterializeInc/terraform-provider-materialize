package materialize

import (
	"reflect"
	"testing"
)

func TestAreEqual(t *testing.T) {
	o := areEqual(map[string]interface{}{"name": "name", "alias": "alias"}, map[string]interface{}{"name": "name", "alias": "alias"})
	if !o {
		t.Fatalf("Expected areEqual to be equal")
	}
}

func TestAreUnequal(t *testing.T) {
	o := areEqual(
		map[string]interface{}{"name": "name", "alias": "alias"},
		map[string]interface{}{"name": "diff", "alias": "alias"},
	)
	if o {
		t.Fatalf("Expected areEqual to not be equal")
	}
}

func TestDiffTableStructs(t *testing.T) {
	arr1 := []interface{}{
		map[string]interface{}{"name": "old", "schema_name": "public", "alias": "old", "alias_schema_name": "public"},
		map[string]interface{}{"name": "old_1", "schema_name": "public", "alias": "old_2", "alias_schema_name": "public"},
		map[string]interface{}{"name": "shared", "schema_name": "public", "alias": "shared", "alias_schema_name": "public"},
	}
	arr2 := []interface{}{
		map[string]interface{}{"name": "shared", "schema_name": "public", "alias": "shared", "alias_schema_name": "public"},
		map[string]interface{}{"name": "new", "schema_name": "public", "alias": "new", "alias_schema_name": "public"},
	}
	o := DiffTableStructs(arr1, arr2)
	e := []TableStruct{
		{Name: "old", SchemaName: "public", Alias: "old", AliasSchemaName: "public"},
		{Name: "old_1", SchemaName: "public", Alias: "old_2", AliasSchemaName: "public"},
	}
	if !reflect.DeepEqual(o, e) {
		t.Fatalf("Expect %s %s to be equal", o, e)
	}
}
