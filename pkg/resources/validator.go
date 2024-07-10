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

func validateServiceUsername(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	forbiddenPrefixes := []string{"mz_", "pg_", "external_"}

	// Check for "@" character
	if strings.Contains(v, "@") {
		errs = append(errs, fmt.Errorf("%q must not contain '@', got: %s", key, v))
	}

	// Check for forbidden prefixes
	for _, prefix := range forbiddenPrefixes {
		if strings.HasPrefix(v, prefix) {
			errs = append(errs, fmt.Errorf("%q must not start with %s, got: %s", key, prefix, v))
			break
		}
	}

	return warns, errs
}

// validateKafkaTopicConfigStringMap validates a map of string to string for Kafka topic configs
// It checks that the keys are not empty and that the values are not empty strings
func validateKafkaTopicConfigStringMap(val interface{}, key string) (warns []string, errs []error) {
	v, ok := val.(map[string]interface{})
	if !ok {
		errs = append(errs, fmt.Errorf("%q must be a map of string to string", key))
		return
	}
	for k, vv := range v {
		if k == "" {
			errs = append(errs, fmt.Errorf("%q contains an empty key", key))
		}
		s, ok := vv.(string)
		if !ok {
			errs = append(errs, fmt.Errorf("%q.%s must be a string", key, k))
		} else if s == "" {
			errs = append(errs, fmt.Errorf("%q.%s cannot be an empty string", key, k))
		}
	}
	return
}
