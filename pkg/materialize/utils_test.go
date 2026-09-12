package materialize

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQualifiedName(t *testing.T) {
	r := require.New(t)
	q := QualifiedName("database", "schema", "resource")
	r.Equal(q, `"database"."schema"."resource"`)

	rs := require.New(t)
	qs := QualifiedName("database", "schema")
	rs.Equal(qs, `"database"."schema"`)
}

// StringArray parses the Postgres text[] wire format. pgx v5 replaced the
// pgtype.TextArray API this used to lean on, so these pin the formats we
// actually read back from Materialize, quoting and empty grantees included.
func TestStringArrayWireFormats(t *testing.T) {
	cases := []struct {
		name string
		src  interface{}
		want []string
	}{
		{"simple []byte", []byte(`{a,b,c}`), []string{"a", "b", "c"}},
		{"simple string", `{a,b,c}`, []string{"a", "b", "c"}},
		{"empty array", []byte(`{}`), []string{}},
		{"nil", nil, nil},
		{"quoted with spaces", []byte(`{"a b","c,d"}`), []string{"a b", "c,d"}},
		{"real privileges", []byte(`{s1=arwd/s1,u1=UC/u18,=UC/s1}`), []string{"s1=arwd/s1", "u1=UC/u18", "=UC/s1"}},
		{"embedded quote", []byte(`{"he said \"hi\""}`), []string{`he said "hi"`}},
		{"availability zones", []byte(`{use1-az1,use1-az2}`), []string{"use1-az1", "use1-az2"}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var a StringArray
			if err := a.Scan(tt.src); err != nil {
				t.Fatalf("Scan(%v): %s", tt.src, err)
			}
			if !reflect.DeepEqual([]string(a), tt.want) {
				t.Fatalf("got %#v want %#v", []string(a), tt.want)
			}
		})
	}
}

func TestStringArrayRoundTrip(t *testing.T) {
	for _, in := range [][]string{
		{"a", "b"},
		{"s1=arwd/s1", "=UC/s1"},
		{"with space", "with,comma", `with"quote`},
		{},
	} {
		v, err := StringArray(in).Value()
		if err != nil {
			t.Fatalf("Value(%v): %s", in, err)
		}
		var back StringArray
		if err := back.Scan(v); err != nil {
			t.Fatalf("Scan(%v): %s", v, err)
		}
		if !reflect.DeepEqual([]string(back), in) {
			t.Fatalf("round trip: got %#v want %#v (wire %q)", []string(back), in, v)
		}
	}
}
