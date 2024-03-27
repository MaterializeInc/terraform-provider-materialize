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

func TestRoleParameterCreateAndUpdate(t *testing.T) {
	r := require.New(t)

	inRoleParameter := map[string]interface{}{
		"role_name":      "test_role",
		"variable_name":  "transaction_isolation",
		"variable_value": "read committed",
	}

	d := schema.TestResourceDataRaw(t, RoleParameter().Schema, inRoleParameter)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER ROLE "test_role" SET "transaction_isolation" TO 'read committed';`).WillReturnResult(sqlmock.NewResult(0, 1))

		if err := roleParameterCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		// Simulate update by changing the variable value
		d.Set("variable_value", "serializable")
		mock.ExpectExec(`ALTER ROLE "test_role" SET "transaction_isolation" TO 'serializable';`).WillReturnResult(sqlmock.NewResult(0, 1))

		if err := roleParameterUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRoleParameterDelete(t *testing.T) {
	r := require.New(t)

	inRoleParameter := map[string]interface{}{
		"role_name":     "test_role",
		"variable_name": "transaction_isolation",
	}

	d := schema.TestResourceDataRaw(t, RoleParameter().Schema, inRoleParameter)
	d.SetId("test_role-transaction_isolation")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER ROLE "test_role" RESET "transaction_isolation";`).WillReturnResult(sqlmock.NewResult(0, 1))

		if err := roleParameterDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
