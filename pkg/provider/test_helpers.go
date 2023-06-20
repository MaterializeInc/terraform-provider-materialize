package provider

import (
	"math/rand"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
)

func randomPrivilege(objectType string) string {
	p := materialize.ObjectPermissions[objectType].Permissions
	n := rand.Intn(len(p))
	i := p[n]
	return materialize.Permissions[i]
}
