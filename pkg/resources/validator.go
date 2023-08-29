package resources

import (
	"fmt"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func validPrivileges(objType string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return warnings, errors
		}

		allowedP := materialize.ObjectPermissions[objType].Permissions
		for _, p := range allowedP {

			privilege := materialize.Permissions[p]

			if v == privilege {
				return warnings, errors
			}
		}

		var f []string
		for _, p := range materialize.ObjectPermissions[objType].Permissions {
			f = append(f, fmt.Sprintf(`'%s'`, materialize.Permissions[p]))
		}
		fs := strings.Join(f[:], ", ")

		errors = append(errors, fmt.Errorf("expected %s to be one of (%v), got '%s'", k, fs, v))
		return warnings, errors
	}
}
