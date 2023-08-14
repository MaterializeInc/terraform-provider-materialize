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
		map[string]interface{}{"name": "old", "alias": "old"},
		map[string]interface{}{"name": "old_1", "alias": "old_2"},
		map[string]interface{}{"name": "shared", "alias": "shared"},
	}
	arr2 := []interface{}{
		map[string]interface{}{"name": "shared", "alias": "shared"},
		map[string]interface{}{"name": "new", "alias": "new"},
	}
	o := DiffTableStructs(arr1, arr2)
	e := []TableStruct{
		{Name: "old", Alias: "old"},
		{Name: "old_1", Alias: "old_2"},
	}
	if !reflect.DeepEqual(o, e) {
		t.Fatalf("Expect %s %s to be equal", o, e)
	}
}
