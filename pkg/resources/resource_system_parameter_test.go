package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestSystemParameterCreateAndUpdate(t *testing.T) {
	r := require.New(t)

	inSystemParameter := map[string]interface{}{
		"name":  "max_connections",
		"value": "100",
	}

	d := schema.TestResourceDataRaw(t, SystemParameter().Schema, inSystemParameter)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER SYSTEM SET "max_connections" TO '100';`).WillReturnResult(sqlmock.NewResult(0, 1))

		if err := systemParameterCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		// Simulate update by changing the value
		d.Set("value", "200")
		mock.ExpectExec(`ALTER SYSTEM SET "max_connections" TO '200';`).WillReturnResult(sqlmock.NewResult(0, 1))

		if err := systemParameterUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSystemParameterRead(t *testing.T) {
	r := require.New(t)

	inSystemParameter := map[string]interface{}{
		"name": "max_connections",
	}

	d := schema.TestResourceDataRaw(t, SystemParameter().Schema, inSystemParameter)
	d.SetId("max_connections")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectQuery(`SHOW "max_connections";`).WillReturnRows(sqlmock.NewRows([]string{"Value"}).AddRow("100"))

		if err := systemParameterRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		r.Equal("100", d.Get("value").(string))
	})
}

func TestSystemParameterDelete(t *testing.T) {
	r := require.New(t)

	inSystemParameter := map[string]interface{}{
		"name": "max_connections",
	}

	d := schema.TestResourceDataRaw(t, SystemParameter().Schema, inSystemParameter)
	d.SetId("max_connections")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER SYSTEM RESET "max_connections";`).WillReturnResult(sqlmock.NewResult(0, 1))

		if err := systemParameterDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
